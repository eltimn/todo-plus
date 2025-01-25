package routes

import (
	"context"
	"eltimn/todo-plus/models"
	"eltimn/todo-plus/pkg/errs"
	"eltimn/todo-plus/pkg/router"
	"eltimn/todo-plus/web/pages"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	datastar "github.com/starfederation/datastar/sdk/go"
	"github.com/thanhpk/randstr"
)

type ContextKey string

const ContextUserKey ContextKey = "user"
const ContextSessionIdKey ContextKey = "sessionId"
const ContextSessionKey ContextKey = "session"
const ContextIsLoggedInKey ContextKey = "isLoggedIn"
const ContextNonceKey ContextKey = "nonce"

type RouteEnv struct {
	Users    *models.UserModel
	Todos    *models.TodoModel
	Sessions *models.SessionModel

	IsSecure   bool
	AssetsPath string
}

func Routes(env *RouteEnv) *router.Router {
	userEnv := &userEnv{
		users:    env.Users,
		sessions: env.Sessions,
		isSecure: env.IsSecure,
	}

	rtr := router.NewRouter(router.WithErrorHandler(handleHttpError))
	rtr.Use(sessionMiddleware(userEnv))
	rtr.Use(cspMiddleware())

	// serve static files
	fs := http.FileServer(http.Dir(env.AssetsPath))
	rtr.ServeMux.Handle("GET /assets/", http.StripPrefix("/assets/", fs))

	userRoutes(rtr, userEnv)
	todoRoutes(rtr, env.Todos)
	counterRoutes(rtr)

	rtr.Get("/hello", helloHandler)
	rtr.Get("/now", nowHandler)
	rtr.Get("/quiz", quizSSEHandler)
	rtr.Get("/error", func(rw http.ResponseWriter, req *http.Request) error {
		return fmt.Errorf("this is only a test error")
	})

	// Handles the home page and all non-matches
	rtr.ServeMux.Handle("/", homeHandler(userEnv))

	return rtr
}

// This is written as a regular http.HandlerFunc so it can be used as a catch-all route to handle 404s.
func homeHandler(env *userEnv) http.Handler {
	// This middleware is not called otherwise. Probably because ServeMux is called directly.
	mid := sessionMiddleware(env)

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		nonce := contextNonce(req)

		slog.Debug("URL", slog.String("url", req.URL.Path))
		if req.URL.Path != "/" || req.Method != http.MethodGet {
			handleHttpError(rw, req, errs.NotFoundError)
			return
		}

		usr := contextUser(req)

		err := pages.HomePage(usr, nonce).Render(req.Context(), rw)
		if err != nil {
			handleHttpError(rw, req, err)
			return
		}
	})

	return mid(next)
}

func helloHandler(rw http.ResponseWriter, req *http.Request) error {
	usr := contextUser(req)
	nonce := contextNonce(req)

	return pages.Hello(usr, nonce).Render(req.Context(), rw)
}

func nowHandler(rw http.ResponseWriter, req *http.Request) error {
	usr := contextUser(req)
	nonce := contextNonce(req)
	return pages.NowPage(usr, time.Now(), nonce).Render(req.Context(), rw)
}

func quizSSEHandler(w http.ResponseWriter, req *http.Request) error {
	time.Sleep(1 * time.Second) // the the loading spinner
	// Creates a new `ServerSentEventGenerator` instance.
	sse := datastar.NewSSE(w, req)

	// Merges HTML fragments into the DOM.
	err := sse.MergeFragments(`<div id="question">What do you put in a toaster?</div>`)
	if err != nil {
		return err
	}

	// Merges signals into the signals.
	return sse.MergeSignals([]byte(`{response: '', answer: 'bread'}`))
}

func handleHttpError(rw http.ResponseWriter, req *http.Request, err error) {
	slog.Error(err.Error(), errs.ErrAttr(err))

	// Check if the error was an HttpError or a regular error.
	var e errs.HttpError
	switch err := err.(type) {
	case nil:
		e = errs.NewHttpError(fmt.Errorf("nil error"))
	case errs.HttpError: // already an HttpError
		e = err
	default:
		e = errs.NewHttpError(err)
	}

	rw.WriteHeader(e.StatusCode)

	// isHxRequest := req.Header.Get("HX-Request")
	// if isHxRequest == "true" {
	// 	ErrorPartial(e).Render(req.Context(), rw)
	// } else {
	// 	usr := contextUser(req)
	// 	ErrorPage(usr, e).Render(req.Context(), rw)
	// }

	isDatastarRequest := req.Header.Get("Datastar-Request")
	if isDatastarRequest == "true" {
		// ErrorPartial(e).Render(req.Context(), rw)
		sse := datastar.NewSSE(rw, req)
		err := sse.MergeFragmentTempl(ErrorPartial(e))
		if err != nil {
			slog.Error("Error rendering datastart ErrorPartial", errs.ErrAttr(err))
		}
	} else {
		usr := contextUser(req)
		nonce := contextNonce(req)
		err := ErrorPage(usr, e, nonce).Render(req.Context(), rw)
		if err != nil {
			slog.Error("Error rendering ErrorPage", errs.ErrAttr(err))
		}
	}
}

func cspMiddleware() router.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			// "report-uri https://example.com/_csp"
			nonce := randstr.String(24)
			csp := fmt.Sprintf("object-src 'none'; script-src 'strict-dynamic' 'nonce-%s'; base-uri 'self'", nonce)
			rw.Header().Set("Content-Security-Policy", csp)
			slog.Debug("csp", slog.String("middleware", "header set"))

			ctx := context.WithValue(req.Context(), ContextNonceKey, nonce)

			next.ServeHTTP(rw, req.WithContext(ctx))
		})
	}
}

func contextNonce(req *http.Request) string {
	nonce, ok := req.Context().Value(ContextNonceKey).(string)
	if !ok {
		return ""
	}
	return nonce
}
