package gotcha

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	espeak "github.com/djangulo/go-espeak"
	"github.com/djangulo/go-espeak/wav"
	gostorage "github.com/djangulo/go-storage"

	"github.com/djangulo/gotcha/draw"
)

type Generator interface {
	GenerateChallenge(noise float64) (*Captcha, error)
}

type Source int

const (
	// Math randomly generate word math questions "six plus 2", "10 minus five"
	Math Source = 1 << iota
	// Random randomly generate 5 alphanumeric strings
	Random
	// QuestionBank use a provided question bank of the type
	// "what color are the smurfs", ans "blue"
	QuestionBank
)

// Storer interface for persistent storage.
type Storer interface {
	Create(c *Captcha) error
	Get(id uint32) (*Captcha, error)
	Update(id uint32, c *Captcha) error
	Delete(id uint32) error
	// GC garbage collects expired captchas. Needs to run in a goroutine to clean
	// expired captchas.
	GC() error
}
type Challenger interface {
	// Gen generates a Challenge based on parameters in Gen.
	Gen(ctx context.Context) (*Captcha, error)
	Refresh(captchaID uint32) (*Captcha, error)
	Status(captchaID uint32) bool
	// GC garbage collects expired captchas. Needs to run in a goroutine to clean
	// expired captchas. It call the store's GC method every unit.
	GC(unit time.Duration, err chan<- error, cancel <-chan struct{})
}

type Manager struct {
	Challenger
	// Sources sources the manager will use for challenges.
	Sources Source
	// Languages the manager will accept and generate challenges in.
	Languages []string
	// Bank question bank to use.
	Bank *Bank
	// Math set of math symbols to use.
	Math                *Symbols
	Store               Storer
	FileStorage         gostorage.Driver
	defaultExpiry       time.Duration
	lifetimeAfterPassed time.Duration
	publicURL           string
	noGzip              bool
	clients             []string
	mountpoint          string
}

// func (m *Manager) Gen(lang string) (*Captcha, error) {

// }

var (
	// ErrNoSources sources not set, select one of Random, Math or QuestionBank.
	ErrNoSources = errors.New("sources not set, select one of Random, Math or QuestionBank")
	// ErrLangEmpty lang is empty.
	ErrLangEmpty = errors.New("lang is empty")
)

func (m *Manager) Gen(ctx context.Context) (c *Captcha, err error) {
	lang := "en"
	if l, ok := ctx.Value(Language).(string); ok {
		lang = l
	}
	var exp = m.defaultExpiry
	if e, ok := ctx.Value(Expiry).(time.Duration); ok {
		exp = e
	}
	if m.Sources == 0 {
		err = ErrNoSources
		return
	}

	switch v := m.Sources; {
	case v == Math|QuestionBank|Random:
		coin := rand.Float64()
		if coin >= 0 && coin < 0.33 {
			c, err = m.mathChallenge(ctx, lang, exp)
		} else if coin >= 0.33 && coin < 0.66 {
			c, err = m.randomQuery(ctx, lang, exp)
		} else {
			c, err = m.qAndAChallenge(ctx, lang, exp)
		}
	case v == Math|QuestionBank:
		coin := rand.Float64()
		if coin > 0.5 {
			c, err = m.mathChallenge(ctx, lang, exp)
		} else {
			c, err = m.qAndAChallenge(ctx, lang, exp)
		}
	case v == Random|QuestionBank:
		coin := rand.Float64()
		if coin > 0.5 {
			c, err = m.qAndAChallenge(ctx, lang, exp)
		} else {
			c, err = m.randomQuery(ctx, lang, exp)
		}
	case v == QuestionBank:
		c, err = m.qAndAChallenge(ctx, lang, exp)
	case v == Math:
		c, err = m.mathChallenge(ctx, lang, exp)
	case v == Random:
		c, err = m.randomQuery(ctx, lang, exp)
	case v == Math|Random:
		fallthrough
	default:
		coin := rand.Float64()
		if coin > 0.5 {
			c, err = m.mathChallenge(ctx, lang, exp)
		} else {
			c, err = m.randomQuery(ctx, lang, exp)
		}
	}
	if err != nil {
		return nil, err
	}
	create := true
	if v, ok := ctx.Value(createNew).(bool); ok {
		create = v
	}
	if create {
		c.ID = rand.Uint32()
		err = m.Store.Create(c)
		if err != nil {
			return nil, err
		}
	}
	return
}

func (m *Manager) Refresh(captchaID uint32) (*Captcha, error) {
	c, err := m.Store.Get(captchaID)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	ctx, err = AddToContext(ctx,
		Language, c.Lang,
		Expiry, c.Expiry,
		createNew, false,
	)
	if err != nil {
		return nil, err
	}
	c, err = m.Gen(ctx)
	if err != nil {
		return nil, err
	}
	c.ID = captchaID
	err = m.Store.Update(captchaID, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// GC needs to be run in a goroutine to clean expired captcahs. It'll start a
// time.Ticker every unit and periodically call the store GC method.
func (m *Manager) GC(unit time.Duration, errChan chan<- error, cancel <-chan struct{}) {

	tick := time.Tick(1 * unit)
	for {
		select {
		case <-tick:
			if err := m.Store.GC(); err != nil {
				errChan <- err
			}
		case <-cancel:
			return
		}
	}
}

// NewManager returns a default manager modified with opts. Panics on error.
func NewManager(opts ...Option) *Manager {
	m := DefaultManager
	for _, opt := range opts {
		opt(m)
	}
	return m
}

type Challenge interface {
	// Match check whether the answer passes the test.
	Match(answer string) bool
}

type Captcha struct {
	// ID of this Captcha.
	ID uint32 `json:"id,omitempty"`
	// Img url to image file on server.
	Image string `json:"image-url,omitempty"`
	// Passed status of this captcha.
	Passed   bool   `json:"passed"`
	ClientID string `json:"client-id"`
	// Lang language.
	Lang string `json:"language,omitempty"`
	// Audio url to audio file on server.
	Audio string `json:"audio-url,omitempty"`
	// Question the question posed in the captcha.
	Question string `json:"question,omitempty"`
	// Ans slice of acceptable answers.
	Answers []string `json:"answers,omitempty"`
	// Expiry when this Captcha is invalid and needs to be refreshed.
	Expiry time.Time `json:"expiry,omitempty"`
}

func (q *Captcha) Match(ans string) bool {
	for _, a := range q.Answers {
		if a == ans {
			return true
		}
	}
	return false
}

// func (q *Captcha) HTML() template.HTML {
// 	var b strings.Builder
// 	b.WriteString(fmt.Sprintf(`<div class="captcha" data-gotcha-id="%d" style="display: block; width:12em;">
// 	<img style="width: 100%%;" alt="captcha image" id="challenge-base64image" src="data:image/jpeg;base64, %s" />
//     <input style="width: 100%%;" type="text" id="challenge-response" name="challenge-response"/>
// </div>
// 	`, q.B64Image))

// 	return template.HTML(b.String())
// }

var (
	defaultLangs = []string{"en", "es", "fr"}
	// DefaultManager default settings.
	DefaultManager = &Manager{
		Languages:           defaultLangs,
		Sources:             Math | Random,
		Math:                NewSymbols(defaultLangs...),
		Bank:                NewBank(defaultLangs...),
		defaultExpiry:       10 * time.Minute,
		lifetimeAfterPassed: 2 * time.Minute,
		Store:               DefaultStore,
	}
)

func init() {
	var data []byte
	var err error
	var r *bytes.Reader

	panicErr := func(e error) {
		if e != nil {
			panic(e)
		}
	}

	for _, lang := range DefaultManager.Languages {
		data, err = Asset(filepath.Join("locale", lang, "symbols.json"))
		panicErr(err)
		r = bytes.NewReader(data)
		var sym []*MathSymbol
		err = json.NewDecoder(r).Decode(&sym)
		panicErr(err)
		DefaultManager.Math.Values[lang] = sym

		data, err, r = nil, nil, nil

		data, err = Asset(filepath.Join("locale", lang, "bank.json"))
		panicErr(err)
		r = bytes.NewReader(data)
		var captchas []*Captcha
		err = json.NewDecoder(r).Decode(&captchas)
		panicErr(err)
		DefaultManager.Bank.Values[lang] = captchas
	}
}

// type RandomQuery struct {
// 	Q        string
// 	B64Image string
// }

func (m *Manager) randomQuery(ctx context.Context, lang string, exp time.Duration) (*Captcha, error) {
	str := randomString(6)
	c := &Captcha{
		Question: str,
		Answers:  []string{str},
		Expiry:   time.Now().Add(exp),
	}
	var err error
	c.Image, c.Audio, err = m.getMedia(ctx, c.ID, lang, str)
	if err != nil {
		return nil, err
	}
	return c, nil
}

type fPool [][]draw.FuzzFactory

func (fp fPool) rand() []draw.FuzzFactory {
	return fp[rand.Intn(len(fp))]
}

var fuzzerPool = fPool{
	{draw.Bands, draw.RandomLines},
	{draw.Bands, draw.RandomCircles},
	{draw.Bands, draw.ConcentricCircles},
	{draw.RandomLines, draw.ConcentricCircles},
	{draw.RandomLines, draw.RandomCircles},
	{draw.RandomCircles, draw.ConcentricCircles},
	// {draw.Bands},
	// {draw.ConcentricCircles},
	// {draw.RandomCircles},
	// {draw.RandomLines},
}

func (m *Manager) getMedia(
	ctx context.Context,
	id uint32,
	lang string,
	q string,
) (imgURL string, audioURL string, err error) {
	var fuzzers = make([]draw.Fuzzer, 0)
	for _, f := range fuzzerPool.rand() {
		fuzzers = append(fuzzers, f(ctx))
	}
	img := draw.Gen(
		q,
		draw.Inconsolata(ctx),
		fuzzers...)

	var buf bytes.Buffer
	png.Encode(&buf, img)

	imgURL, err = m.FileStorage.AddFile(
		&buf,
		filepath.Join(strconv.Itoa(int(id)), "image.png"),
	)
	if err != nil {
		return "", "", err
	}
	buf.Reset()

	v, err := espeak.VoiceFromSpec(&espeak.Voice{Languages: lang})
	if err != nil {
		return "", "", err
	}
	audioSamples, err := espeak.GenSamples(q, v, nil)
	if err != nil {
		return "", "", err
	}
	wav.NewWriter(&buf, espeak.SampleRate()).WriteSamples(audioSamples)
	audioURL, err = m.FileStorage.AddFile(
		&buf,
		filepath.Join(strconv.Itoa(int(id)), "audio.wav"),
	)
	if err != nil {
		return "", "", err
	}

	return
}

// Bank contains arrays a map[string][]*Captcha, under each language.
// The captchas only contain questions and answers, their Audio and
// image fields will not be populated until it is requested.
type Bank struct {
	Values map[string][]*Captcha
}

// NewBank initializes a *Bank with langs.
func NewBank(langs ...string) *Bank {
	var b = &Bank{Values: make(map[string][]*Captcha)}
	for _, lang := range langs {
		b.Values[lang] = make([]*Captcha, 0)
	}
	return b
}

func (m *Manager) qAndAChallenge(ctx context.Context, lang string, exp time.Duration) (*Captcha, error) {
	if len(m.Bank.Values) == 0 || len(m.Bank.Values[lang]) == 0 {
		return nil, fmt.Errorf("bank is empty")
	}
	c := m.Bank.Values[lang][rand.Intn(len(m.Bank.Values[lang]))]
	c.Expiry = time.Now().Add(exp)

	var err error
	c.Image, c.Audio, err = m.getMedia(ctx, c.ID, lang, c.Question)
	if err != nil {
		return nil, err
	}
	return c, nil
}

type MathSymbol struct {
	Human  string `json:"human"`
	Symbol string `json:"symbol"`
}

func (m *MathSymbol) getValue() (int, error) {
	return strconv.Atoi(m.Symbol)
}

// Symbols holds MathSymbol definitions under each language.
// e.g. Symbols.Values["en"].
type Symbols struct {
	Values map[string][]*MathSymbol
}

// NewSymbols initializes a *Symbols with langs.
func NewSymbols(langs ...string) *Symbols {
	var s = &Symbols{Values: make(map[string][]*MathSymbol)}
	for _, lang := range langs {
		s.Values[lang] = make([]*MathSymbol, 0)
	}
	return s
}

// String always return the symbol be parsed.
func (m *MathSymbol) String() string {
	return m.Symbol
}

// HumanOrSymbol 50/50 chance of either.
func (m *MathSymbol) HumanOrSymbol() string {
	c := rand.Float64()
	if c > 0.5 {
		return m.Human
	}
	return m.Symbol
}

func (m *Manager) mathChallenge(ctx context.Context, lang string, exp time.Duration) (*Captcha, error) {
	var a, b, sym *MathSymbol
	ra, rb, rsym := "0", "0", ""
	for ra == "0" {
		ra = strconv.Itoa(rand.Intn(10))
	}
	for rb == "0" {
		rb = strconv.Itoa(rand.Intn(10))
	}
	rsym = [...]string{"+", "-", "×", "÷"}[rand.Intn(4)]
	for _, entry := range m.Math.Values[lang] {
		if entry.Symbol == ra {
			a = entry
		}
		if entry.Symbol == rb {
			b = entry
		}
		if entry.Symbol == rsym {
			sym = entry
		}
		if a != nil && b != nil && sym != nil {
			break
		}
	}

	av, err := a.getValue()
	if err != nil {
		return nil, err
	}
	bv, err := b.getValue()
	if err != nil {
		return nil, err
	}
	// bigest number on the left, for simpler division and
	// substraction
	if av < bv {
		a, b = b, a
	}

	expr := fmt.Sprintf("%s %s %s", a, sym, b)
	q := fmt.Sprintf("%s %s %s", a.HumanOrSymbol(), sym.HumanOrSymbol(), b.HumanOrSymbol())

	numberAns, err := parseExpr(expr)
	if err != nil {
		return nil, err
	}
	var stringAns string
	for _, entry := range m.Math.Values[lang] {
		if entry.Symbol == numberAns {
			stringAns = entry.Human
			break
		}
	}
	c := &Captcha{
		Question: q,
		Answers:  []string{numberAns, stringAns},
		Expiry:   time.Now().Add(exp),
	}

	c.Image, c.Audio, err = m.getMedia(ctx, c.ID, lang, q)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// AddToContext is a helper function that adds the values in keyValuePair to
// ctx. Keys are even indexes (0 is considered even), values are odd indexes.
func AddToContext(ctx context.Context, keyValuePair ...interface{}) (context.Context, error) {
	if len(keyValuePair)%2 != 0 {
		return nil, fmt.Errorf("keyValuePair is not even")
	}
	for i := 0; i < len(keyValuePair); i += 2 {
		ctx = context.WithValue(ctx, keyValuePair[i], keyValuePair[i+1])
	}
	return ctx, nil

}

// func mathCaptcha(
// 	symbolSet *Symbols,
// 	lang string,
// 	noise float64,
// 	storage gostorage.Driver,
// ) (*Captcha, error) {
// 	var a, b, sym *MathSymbol
// 	ra, rb, rsym := "0", "0", ""
// 	for ra == "0" {
// 		ra = strconv.Itoa(rand.Intn(10))
// 	}
// 	for rb == "0" {
// 		rb = strconv.Itoa(rand.Intn(10))
// 	}
// 	rsym = [...]string{"+", "-", "×", "÷"}[rand.Intn(4)]
// 	for _, entry := range symbolSet.Values[lang] {
// 		if entry.Symbol == ra {
// 			a = entry
// 		}
// 		if entry.Symbol == rb {
// 			b = entry
// 		}
// 		if entry.Symbol == rsym {
// 			sym = entry
// 		}
// 		if a != nil && b != nil && sym != nil {
// 			break
// 		}
// 	}

// 	av, err := a.getValue()
// 	if err != nil {
// 		return nil, err
// 	}
// 	bv, err := b.getValue()
// 	if err != nil {
// 		return nil, err
// 	}
// 	// bigest number on the left, for simpler division and
// 	// substraction
// 	if av < bv {
// 		a, b = b, a
// 	}

// 	expr := fmt.Sprintf("%s %s %s", a, sym, b)
// 	strExpr := fmt.Sprintf("%s %s %s", a.Human, sym.Human, b.Human)
// 	espeak.TextToSpeech()
// 	q := fmt.Sprintf("%s %s %s", a.HumanOrSymbol(), sym.HumanOrSymbol(), b.HumanOrSymbol())

// 	numberAns, err := parseExpr(expr)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var stringAns string
// 	for _, entry := range *s {
// 		if entry.Symbol == numberAns {
// 			stringAns = entry.Human
// 			break
// 		}
// 	}
// 	img := genImage(q, noise)
// 	return &Captcha{
// 		Q:   q,
// 		Ans: []string{numberAns, stringAns},
// 		Img: img,
// 	}, nil

// }

func parseExpr(expr string) (string, error) {
	split := strings.Split(expr, " ")
	var err error
	var a, b int
	a, err = strconv.Atoi(split[0])
	if err != nil {
		return "0", err
	}
	b, err = strconv.Atoi(split[2])
	if err != nil {
		return "0", err
	}
	switch split[1] {
	case "+":
		return strconv.Itoa(a + b), nil
	case "-":
		if a > b {
			return strconv.Itoa(a - b), nil
		}
		return strconv.Itoa(b - a), nil
	case "×":
		return strconv.Itoa(a * b), nil
	case "÷":
		if a > b {
			return strconv.Itoa(a / b), nil
		}
		return strconv.Itoa(b / a), nil
	default:
		return "0", fmt.Errorf("unknown operation: %s", split[1])
	}
}

const (
	chars      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bitsNeeded = 6 // len(chars[:63]) = 62 = 0b111110
	mask       = 1<<bitsNeeded - 1
	max        = 63 / bitsNeeded
)

// randomString generate a random string of length n
// Modified from https://stackoverflow.com/a/31832326
func randomString(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, rand.Int63(), max; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), max
		}
		if idx := int(cache & mask); idx < len(chars) {
			b[i] = chars[idx]
			i--
		}
		cache >>= bitsNeeded
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

func readJSON(path string, target interface{}) error {
	fh, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fh.Close()

	err = json.NewDecoder(fh).Decode(target)
	if err != nil {
		return err
	}
	return nil
}

var DefaultStore Storer = &defaultStore{
	captchas: make(map[uint32]*Captcha),
}

type defaultStore struct {
	sync.Mutex
	captchas map[uint32]*Captcha
}

func (ds *defaultStore) Create(c *Captcha) error {
	ds.Lock()
	defer ds.Unlock()
	ds.captchas[c.ID] = c
	return nil
}

func (ds *defaultStore) Get(id uint32) (*Captcha, error) {
	ds.Lock()
	defer ds.Unlock()
	if c, ok := ds.captchas[id]; ok {
		return c, nil
	}

	return nil, fmt.Errorf("not found")
}

func (ds *defaultStore) Update(id uint32, c *Captcha) error {
	ds.Lock()
	defer ds.Unlock()

	ds.captchas[id] = c
	return nil
}

func (ds *defaultStore) Delete(id uint32) error {
	ds.Lock()
	defer ds.Unlock()

	ds.captchas[id] = nil
	delete(ds.captchas, id)
	return nil
}

func (ds *defaultStore) GC() error {
	ds.Lock()
	defer ds.Unlock()

	now := time.Now()
	for id, c := range ds.captchas {
		if now.After(c.Expiry) {
			err := ds.Delete(id)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
