package router

import (
	"fmt"
	"net/http"
	"slices"
	"sync/atomic"
)

// router options
type HttpErrorHandler func(http.ResponseWriter, *http.Request, error)
type RouterOpts struct {
	HandleError HttpErrorHandler
}

var routerOpts atomic.Pointer[RouterOpts]

func init() {
	routerOpts.Store(&RouterOpts{HandleError: defaultHttpErrorHandler})
}

// Set the error handler
func SetRouterOpts(opts RouterOpts) {
	routerOpts.Store(&opts)
}

// An HTTP handler that allows for returning errors
type HttpHandler func(http.ResponseWriter, *http.Request) error

// ServeHTTP implements the http.Handler interface for HttpHandler
func (fn HttpHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if err := fn(w, req); err != nil {
		opts := routerOpts.Load()
		opts.HandleError(w, req, err)
	}
}

// The default error handler
func defaultHttpErrorHandler(w http.ResponseWriter, req *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// https://gist.github.com/alexaandru/747f9d7bdfb1fa35140b359bf23fa820
// Router is a simple HTTP router that wraps the standard library's ServeMux
type Middleware func(http.Handler) http.Handler
type Router struct {
	*http.ServeMux
	chain []Middleware
}

func NewRouter(opts *RouterOpts, mx ...Middleware) *Router {
	if opts.HandleError != nil {
		routerOpts.Store(opts)
	}
	return &Router{ServeMux: &http.ServeMux{}, chain: mx}
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

func Mid(i int) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("mid", i, "start")
			next.ServeHTTP(w, r)
			fmt.Println("mid", i, "done")
		})
	}
}
