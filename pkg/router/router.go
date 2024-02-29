package router

import (
	"net/http"
	"slices"
	"sync/atomic"
)

// router config
type HttpErrorHandler func(http.ResponseWriter, *http.Request, error)
type RouterConfig struct {
	HandleError HttpErrorHandler
}

type RouterOption func(*RouterConfig)

func WithErrorHandler(fn HttpErrorHandler) RouterOption {
	return func(cfg *RouterConfig) {
		cfg.HandleError = fn
	}
}

var routerCfg atomic.Pointer[RouterConfig]

func init() {
	routerCfg.Store(&RouterConfig{HandleError: defaultHttpErrorHandler})
}

// An HTTP handler that allows for returning errors
type HttpHandler func(http.ResponseWriter, *http.Request) error

// ServeHTTP implements the http.Handler interface for HttpHandler
func (fn HttpHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if err := fn(w, req); err != nil {
		opts := routerCfg.Load()
		opts.HandleError(w, req, err)
	}
}

// The default error handler
func defaultHttpErrorHandler(w http.ResponseWriter, req *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// https://gist.github.com/alexaandru/747f9d7bdfb1fa35140b359bf23fa820
type Middleware func(http.Handler) http.Handler

// Router is a simple HTTP router that wraps the standard library's ServeMux
type Router struct {
	*http.ServeMux
	chain []Middleware
}

func NewRouter(opts ...RouterOption) *Router {

	ro := RouterConfig{
		HandleError: defaultHttpErrorHandler,
	}
	for _, opt := range opts {
		opt(&ro)
	}

	routerCfg.Store(&ro)

	return &Router{ServeMux: &http.ServeMux{}, chain: []Middleware{}}
}

func (r *Router) Use(mx ...Middleware) {
	r.chain = append(r.chain, mx...)
}

func (r *Router) Group(fn func(r *Router)) {
	fn(&Router{ServeMux: r.ServeMux, chain: slices.Clone(r.chain)})
}

func (r *Router) Get(path string, fn HttpHandler, mx ...Middleware) {
	r.handle(http.MethodGet, path, fn, mx)
}

func (r *Router) Patch(path string, fn HttpHandler, mx ...Middleware) {
	r.handle(http.MethodPatch, path, fn, mx)
}

func (r *Router) Post(path string, fn HttpHandler, mx ...Middleware) {
	r.handle(http.MethodPost, path, fn, mx)
}

func (r *Router) Put(path string, fn HttpHandler, mx ...Middleware) {
	r.handle(http.MethodPut, path, fn, mx)
}

func (r *Router) Delete(path string, fn HttpHandler, mx ...Middleware) {
	r.handle(http.MethodDelete, path, fn, mx)
}

func (r *Router) Head(path string, fn HttpHandler, mx ...Middleware) {
	r.handle(http.MethodHead, path, fn, mx)
}

func (r *Router) Options(path string, fn HttpHandler, mx ...Middleware) {
	r.handle(http.MethodOptions, path, fn, mx)
}

func (r *Router) handle(method, path string, fn HttpHandler, mx []Middleware) {
	r.ServeMux.Handle(method+" "+path, r.wrap(fn, mx))
}

func (r *Router) wrap(fn HttpHandler, mx []Middleware) (out http.Handler) {
	out, mx = http.Handler(fn), append(slices.Clone(r.chain), mx...)

	slices.Reverse(mx)

	for _, m := range mx {
		out = m(out)
	}

	return
}
