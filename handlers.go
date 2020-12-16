package gotcha

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/djangulo/go-storage"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

func (m *Manager) Router(ctx context.Context) http.Handler {
	r := chi.NewRouter()
	r.Route(m.mountpoint, func(r chi.Router) {
		r.Get("/new", m.NewCaptcha)
		// r.Post("/register")
		r.Route("/{gotchaID}", func(r chi.Router) {
			r.Use(m.CaptchaCtx)
			r.Post("/check", m.CheckCaptcha)
			r.Post("/refresh", m.CheckCaptcha)
		})
	})
	return r
}

type CaptchaResponse struct {
	*Captcha
}

func NewCaptchaResponse(c *Captcha) *CaptchaResponse {
	return &CaptchaResponse{Captcha: c}
}

func (c *CaptchaResponse) Render(w http.ResponseWriter, r *http.Request) error {
	c.Answers = nil
	return nil
}

func (m *Manager) NewCaptcha(w http.ResponseWriter, r *http.Request) {
	c, err := m.Gen(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := render.Render(w, r, NewCaptchaResponse(c)); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

type CheckRequest struct {
	Answer string `json:"challenge-response"`
}

func (c *CheckRequest) Bind(r *http.Request) error {
	if c.Answer == "" {
		return fmt.Errorf("response is empty")
	}
	return nil
}

type OKResponse struct {
	Status string `json:"Status"`
}

func (ok *OKResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (m *Manager) CheckCaptcha(w http.ResponseWriter, r *http.Request) {
	data := &CheckRequest{}
	var err error
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	captcha := r.Context().Value(CaptchaCtxKey).(*Captcha)
	if !captcha.Match(data.Answer) {
		if err = render.Render(w, r, ErrIncorrectAnswer); err != nil {
			render.Render(w, r, ErrRender(err))
			return
		}
		return
	}
	captcha.Passed = true
	// set expiry to the lifetime after passed, GC will take care of
	// removing the captcha
	captcha.Expiry = time.Now().Add(m.lifetimeAfterPassed)
	err = m.Store.Update(captcha.ID, captcha)
	if err != nil {
		render.Render(w, r, ErrInternalServerError)
		return
	}
	render.Status(r, 200)
	if err := render.Render(w, r, &OKResponse{"OK"}); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

func (m *Manager) RefreshCaptcha(w http.ResponseWriter, r *http.Request) {
	captcha := r.Context().Value(CaptchaCtxKey).(*Captcha)
	var err error
	captcha, err = m.Refresh(captcha.ID)
	if err != nil {
		if err := render.Render(w, r, ErrInternalServerError); err != nil {
			render.Render(w, r, ErrRender(err))
			return
		}
		return
	}
	render.Status(r, 200)
	if err := render.Render(w, r, NewCaptchaResponse(captcha)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

func (m *Manager) CaptchaCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var captcha *Captcha
		var err error

		ctx := r.Context()

		if captchaID := chi.URLParam(r, "captchaID"); captchaID != "" {
			if id, err := strconv.ParseUint(captchaID, 10, 32); err != nil {
				render.Render(w, r, ErrInvalidRequest(err))
				return
			} else {
				captcha, err = m.Store.Get(uint32(id))
			}
		} else {
			render.Render(w, r, ErrNotFound)
			return
		}
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		ctx = context.WithValue(ctx, CaptchaCtxKey, captcha)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

var (
	templatesMu   sync.RWMutex
	templates     map[string]*template.Template
	templateFiles = map[string]struct{}{
		"gotcha.min.js": {},
		"gotcha.js":     {},
	}
)

func getTemplate(filename string) (*template.Template, error) {
	if _, ok := templateFiles[filename]; !ok {
		return nil, fmt.Errorf("unknown template file")
	}
	if t, ok := templates[filename]; ok {
		return t, nil
	}

	templatesMu.Lock()
	defer templatesMu.Unlock()
	b, err := Asset("static/js/" + filename)
	if err != nil {
		return nil, fmt.Errorf("error getting %q: %v", filename, err)
	}
	t, err := template.New(filename).Parse(string(b))
	if err != nil {
		return nil, fmt.Errorf("error parsing %q: %v", filename, err)
	}
	templates[filename] = t
	return t, nil
}

// FileServer serve static files the chi way.
func (m *Manager) fileServer(r chi.Router, fpath string, root http.FileSystem) {
	if strings.ContainsAny(fpath, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if fpath != "/" && fpath[len(fpath)-1] != '/' {
		r.Get(fpath, http.RedirectHandler(fpath+"/", 301).ServeHTTP)
		fpath += "/"
	}
	fpath += "*"

	if !m.noGzip {
		r.Use(middleware.Compress(
			5,
			"text/html",
			"text/css",
			"text/plain",
			"text/javascript",
			"application/javascript",
			"application/x-javascript",
			"application/json",
			"application/atom+xml",
			"application/rss+xml",
			"image/svg+xml",
		))
	}

	r.Get(fpath, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		w.Header().Set("Cache-Control", "max-age:2592000, public")

		pbase := path.Base(r.URL.Path)
		if t, _ := getTemplate(pbase); t != nil {
			var data = struct {
				PublicURL string
			}{
				PublicURL: m.publicURL,
			}
			if err := t.Execute(w, data); err != nil {
				http.Error(w, "error getting file", 500)
				return
			}
			return
		}

		w.Header().Set(
			"Content-Type",
			storage.ResolveContentType(pbase),
		)
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

var (
	ErrNotFound            = &ErrResponse{HTTPStatusCode: 404, StatusText: "Resource not found."}
	ErrInternalServerError = &ErrResponse{HTTPStatusCode: 500, StatusText: "Server error."}
	ErrIncorrectAnswer     = &ErrResponse{HTTPStatusCode: 403, StatusText: "Incorrect answer."}
)
