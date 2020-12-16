package draw

//go:generate go run scripts/createB64fonts.go

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"math/rand"
	"sync"
	"time"
)

const (
	bitsNeeded = 6 // len(chars[:63]) = 62 = 0b111110
	mask       = 1<<bitsNeeded - 1
	max        = 63 / bitsNeeded
)

func init() {
	rand.Seed(time.Now().Unix())
}

// GenBase64 returns the base64 encoding of the image with text in it.
func GenBase64(text string, fd FontDrawer, fuzzers ...Fuzzer) (b64 string) {
	captcha := gen(text, fd, fuzzers...)
	var buf bytes.Buffer
	png.Encode(&buf, captcha)
	b64 = base64.StdEncoding.EncodeToString(buf.Bytes())
	return
}

// Gen returns an image/draw.Image the image with text in it.
func Gen(text string, fd FontDrawer, fuzzers ...Fuzzer) draw.Image {
	return gen(text, fd, fuzzers...)
}

// FontDrawer is the functype that draws fonts into the image. It's
// expected to return the image to work in as the font determines
// the size.
type FontDrawer func(string) draw.Image

var genWG sync.WaitGroup

func gen(text string, fd FontDrawer, fuzzers ...Fuzzer) draw.Image {
	captcha := fd(text)

	for _, fuzzer := range fuzzers {
		genWG.Add(1)
		fn := fuzzer
		go func() {
			fn(captcha)
			defer genWG.Done()
		}()
	}
	genWG.Wait()

	return captcha
}

// Fuzzer helper struct to call each fuzzer function with its own context.
type Fuzzer func(img draw.Image)

// FuzzFactory helper closure that passes a (potentially) unique context to
// the Fuzzer.
type FuzzFactory func(ctx context.Context) Fuzzer

type key int

const (
	// FuzzColor1CtxKey context key for fuzzer functions color (image/color.RGBA).
	FuzzColor1CtxKey key = iota + 1
	// FuzzColor2CtxKey context key for a second color (image/color.RGBA).
	FuzzColor2CtxKey
	// FuzzNoiseCtxKey context key for fuzzer noise (float64).
	FuzzNoiseCtxKey
	// FuzzBandThicknessCtxKey context key for thickness of bands, in pixels (int).
	FuzzBandThicknessCtxKey
	// FuzzSlopeCtxKey context key for line slope. (float64)
	FuzzSlopeCtxKey
	// FontVariantCtxKey key to select one of the font variants. (InconsolataVariant)
	FontVariantCtxKey
	// LineBreak set the linebreak (int).
	LineBreak
	// ScaleBy value to scale fonts by (float64).
	ScaleBy
	// RotateBy degrees to rotate by. Ignored unless RandomRotation explicitly
	// set to false. (int)
	RotateBy
	// RandomRotation whether to apply a random rotation from -35 to 35 degrees. (bool)
	RandomRotation
	// BackgroundColor set the bacground color for mask image. (image/color.RGBA)
	BackgroundColor
)

// RandomLines draws random lines over img. ctx reads noise and color keys.
func RandomLines(ctx context.Context) Fuzzer {

	var col color.RGBA = color.RGBA{192, 192, 192, 192}
	var noise float64 = 0.5
	if c, ok := ctx.Value(FuzzColor1CtxKey).(color.RGBA); ok {
		col = c
	}
	if n, ok := ctx.Value(FuzzNoiseCtxKey).(float64); ok {
		noise = normalizeNoise(n)
	}

	return func(img draw.Image) {
		lines := image.NewRGBA(img.Bounds())

		var wg sync.WaitGroup
		for i := 0; i < int(noise*100); i++ {
			wg.Add(1)
			func() {
				defer wg.Done()
				x0, y0 := rand.Intn(lines.Rect.Max.X), rand.Intn(lines.Rect.Max.Y)
				x1, y1 := rand.Intn(lines.Rect.Max.X), rand.Intn(lines.Rect.Max.Y)
				drawLine(lines, x0, y0, x1, y1, col)
			}()
		}
		wg.Wait()
		draw.Draw(img, img.Bounds(), lines, image.ZP, draw.Over)
	}
}

// RandomCircles draws random grayscale circles over img. ctx reads noise key.
func RandomCircles(ctx context.Context) Fuzzer {
	var noise float64 = 0.5
	if n, ok := ctx.Value(FuzzNoiseCtxKey).(float64); ok {
		noise = normalizeNoise(n)
	}
	var col color.RGBA
	if c, ok := ctx.Value(FuzzColor1CtxKey).(color.RGBA); ok {
		col = c
	} else {
		var k uint8 = uint8(rand.Intn(128))
		col = color.RGBA{k, k, k, k}
	}
	return func(img draw.Image) {

		circles := image.NewRGBA(img.Bounds())

		var wg sync.WaitGroup
		for i := 0; i < int(math.Round(noise*10)); i++ {
			// for i := 0; i < 1; i++ {
			wg.Add(1)
			r := rand.Intn(circles.Rect.Max.Y)
			cx := rand.Intn(circles.Rect.Max.X)
			cy := rand.Intn(circles.Rect.Max.Y)
			go func(r, cx, cy int) {
				defer wg.Done()

				for rr := r; rr > 0; rr-- {
					drawCircle(circles, cx, cy, rr, col)
				}
			}(r, cx, cy)
		}
		wg.Wait()
		draw.Draw(img, img.Bounds(), circles, image.ZP, draw.Over)

	}
}

// ConcentricCircles draws concentric rings over img.
func ConcentricCircles(ctx context.Context) Fuzzer {
	var noise float64 = 0.5
	if n, ok := ctx.Value(FuzzNoiseCtxKey).(float64); ok {
		noise = normalizeNoise(n)
	}
	var col1 = color.RGBA{192, 192, 192, 192}
	var col2 = color.RGBA{128, 128, 128, 128}
	if c, ok := ctx.Value(FuzzColor1CtxKey).(color.RGBA); ok {
		col1 = c
	}
	if c, ok := ctx.Value(FuzzColor2CtxKey).(color.RGBA); ok {
		col2 = c
	}

	var thickness int = 10
	if n, ok := ctx.Value(FuzzBandThicknessCtxKey).(int); ok {
		thickness = n
	}
	return func(img draw.Image) {

		circles := image.NewRGBA(img.Bounds())

		var wg sync.WaitGroup
		for i := 0; i < int(math.Round(noise*20)); i++ {
			// for i := 0; i < 1; i++ {
			wg.Add(1)
			r := rand.Intn(circles.Rect.Max.Y)
			cx := rand.Intn(circles.Rect.Max.X)
			cy := rand.Intn(circles.Rect.Max.Y)
			go func(r, cx, cy int, col1, col2 color.Color) {
				defer wg.Done()

				swtch := true
				var col color.Color
				for j := r; j >= 0; j -= thickness {
					if swtch {
						col = col1
					} else {
						col = col2
					}
					for rr := j; rr >= j-thickness; rr-- {
						drawCircle(circles, cx, cy, rr, col)
					}
					swtch = !swtch
				}
			}(r, cx, cy, col1, col2)
		}
		wg.Wait()
		draw.Draw(img, img.Bounds(), circles, image.ZP, draw.Over)
	}
}

// AccurateBands works just like Bands, but the thickness is on point.
// Is about 1.5x more CPU costly than Bands, and 75x more memory.
func AccurateBands(ctx context.Context) Fuzzer {

	var col1 = color.RGBA{192, 192, 192, 192}
	var col2 = color.RGBA{128, 128, 128, 128}
	if c, ok := ctx.Value(FuzzColor1CtxKey).(color.RGBA); ok {
		col1 = c
	}
	if c, ok := ctx.Value(FuzzColor2CtxKey).(color.RGBA); ok {
		col2 = c
	}

	var slope float64 = 1.0
	if n, ok := ctx.Value(FuzzSlopeCtxKey).(float64); ok {
		slope = n
	}

	if slope > 2.0 {
		slope = 2.0
	} else if slope < -2.0 {
		slope = -2.0
	}
	var thickness int = 10
	if n, ok := ctx.Value(FuzzBandThicknessCtxKey).(int); ok {
		thickness = n
	}

	return func(img draw.Image) {
		bands := image.NewRGBA(img.Bounds())

		width := img.Bounds().Dx()
		var intercept int
		if slope < 0 {
			intercept = -bands.Rect.Max.Y
		} else {
			intercept = 0
		}
		var swtch bool = true

		var wg sync.WaitGroup
		var col color.RGBA

		for j := -thickness; j < bands.Rect.Max.X*2+thickness; j += thickness {
			if swtch {
				col = col1
			} else {
				col = col2
			}
			wg.Add(1)
			go func(b int, col color.RGBA) {
				defer wg.Done()
				for i := 0; i < thickness; i++ {
					for x := -thickness; x <= width; x++ {
						y := -lineEqY(slope, x+i, b) - i
						bands.Set(x+i, y, col)
					}
				}
			}(intercept, col)
			if slope < 0 {
				intercept += thickness
			} else {
				intercept -= thickness
			}
			swtch = !swtch
		}
		wg.Wait()
		draw.Draw(img, img.Bounds(), bands, image.ZP, draw.Over)
	}
}

// Bands draws paralell bands along the image.
func Bands(ctx context.Context) Fuzzer {

	var col1 = color.RGBA{192, 192, 192, 192}
	var col2 = color.RGBA{128, 128, 128, 128}
	if c, ok := ctx.Value(FuzzColor1CtxKey).(color.RGBA); ok {
		col1 = c
	}
	if c, ok := ctx.Value(FuzzColor2CtxKey).(color.RGBA); ok {
		col2 = c
	}

	var slope float64 = -1.0
	if n, ok := ctx.Value(FuzzSlopeCtxKey).(float64); ok {
		slope = normalizeSlope(-n)
	}
	var thickness int = 10
	if n, ok := ctx.Value(FuzzBandThicknessCtxKey).(int); ok {
		thickness = n
	}

	return func(img draw.Image) {
		bands := image.NewRGBA(img.Bounds())

		var wg sync.WaitGroup
		var col color.RGBA

		x := -thickness
		y := 0
		i := 0
		for {
			if i%thickness == 0 {
				if col == col1 {
					col = col2
				} else {
					col = col1
				}
			}
			if y == bands.Rect.Max.Y && x == bands.Rect.Max.X+bands.Rect.Max.Y {
				break
			}

			b := lineEqB(slope, x, y)
			y1 := -lineEqY(slope, x+thickness, b)
			x1 := -lineEqX(slope, -y1, -b)
			wg.Add(1)
			go func(x, y, x1, y2 int, col color.RGBA) {
				defer wg.Done()
				drawLine(bands,
					x,
					y,
					x1,
					y1,
					col,
				)
			}(x, y, x1, y1, col)
			if y == bands.Rect.Max.Y {
				x++
			}
			if y < bands.Rect.Max.Y {
				y++
			}
			i++
		}
		wg.Wait()
		draw.Draw(img, img.Bounds(), bands, image.ZP, draw.Over)
	}
}

// y = mx+b
func lineEqY(m float64, x, b int) int {
	return int(math.Round(m*float64(x) + float64(b)))
}

// x = (y - b)/m
func lineEqX(m float64, y, b int) int {
	return int(math.Round((float64(y) - float64(b)) / m))
}

// b = y - m*x
func lineEqB(m float64, x, y int) int {
	return y - int(math.Round(m*float64(x)))
}

func normalizeNoise(n float64) float64 {
	if n > 1.0 {
		return 1.0
	} else if n < 0.0 {
		return 0.0
	}
	return n
}

func normalizeSlope(m float64) float64 {
	if m > 2.0 {
		return 2.0
	} else if m < -2.0 {
		return -2.0
	}
	return m
}

// midpoint circle algorithm
// see https://stackoverflow.com/a/51627141 and
// https://en.wikipedia.org/wiki/Midpoint_circle_algorithm
func drawCircle(img draw.Image, x0, y0, r int, c color.Color) {
	x, y, dx, dy := r-1, 0, 1, 1
	err := dx - (r * 2)

	for x >= y {
		img.Set(x0+x, y0+y, c)
		img.Set(x0+y, y0+x, c)
		img.Set(x0-y, y0+x, c)
		img.Set(x0-x, y0+y, c)
		img.Set(x0-x, y0-y, c)
		img.Set(x0-y, y0-x, c)
		img.Set(x0+y, y0-x, c)
		img.Set(x0+x, y0-y, c)

		if err <= 0 {
			y++
			err += dy
			dy += 2
		}
		if err > 0 {
			x--
			dx += 2
			err += dx - (r * 2)
		}
	}
}

// Bresenham's line algorithm
// https://en.wikipedia.org/wiki/Bresenham%27s_line_algorithm#All_cases
func drawLine(img draw.Image, x0, y0, x1, y1 int, col color.Color) {
	var dx int = int(math.Abs(float64(x1 - x0)))
	var sx int = -1
	if x0 < x1 {
		sx = 1
	}
	var dy int = -int(math.Abs(float64(y1 - y0)))
	var sy int = -1
	if y0 < y1 {
		sy = 1
	}
	err := dx + dy /* error value e_xy */
	var e2 int
	for {
		img.Set(x0, y0, col)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 = 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}

}
