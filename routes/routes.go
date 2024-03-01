package routes

import (
	"eltimn/todo-plus/models"
	"eltimn/todo-plus/pkg/errs"
	"eltimn/todo-plus/pkg/router"
	"eltimn/todo-plus/web/pages"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type ContextKey string

const ContextUserKey ContextKey = "user"
const ContextSessionIdKey ContextKey = "sessionId"

type RouteEnv struct {
	Users    *models.UserModel
	Todos    *models.TodoModel
	Sessions *models.SessionModel

	IsSecure bool
}

func Routes(env *RouteEnv) *router.Router {
	userEnv := &userEnv{
		users:    env.Users,
		sessions: env.Sessions,
		isSecure: env.IsSecure,
	}

	rtr := router.NewRouter(router.WithErrorHandler(handleHttpError))
	rtr.Use(userSessionMiddleware(userEnv))

	// serve static files
	fs := http.FileServer(http.Dir("./dist/assets"))
	rtr.ServeMux.Handle("GET /assets/", http.StripPrefix("/assets/", fs))

	userRoutes(rtr, userEnv)
	todoRoutes(rtr, env.Todos)

	rtr.Get("/hello", helloHandler)
	rtr.Get("/now", nowHandler)
	rtr.Get("/error", func(rw http.ResponseWriter, req *http.Request) error {
		return fmt.Errorf("this is only a test error")
	})

	// Handles the home page and all non-matches
	rtr.ServeMux.Handle("/", http.HandlerFunc(homeHandler(userEnv)))

	return rtr
}

func helloHandler(rw http.ResponseWriter, req *http.Request) error {
	usr := contextUser(req)
	return pages.Hello(usr).Render(req.Context(), rw)
}

// This is written as a regular http.HandlerFunc so it can be used as a catch-all route to handle 404s.
func homeHandler(env *userEnv) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		slog.Debug("URL", slog.String("url", req.URL.Path))
		if req.URL.Path != "/" || req.Method != http.MethodGet {
			handleHttpError(rw, req, errs.NotFoundError)
			return
		}

		usr, err := getUserFromCookie(req, env)
		if err != nil {
			slog.Debug("User not found", errs.ErrAttr(err))
			handleHttpError(rw, req, err)
			return
		}
		err = pages.HomePage(usr).Render(req.Context(), rw)
		if err != nil {
			handleHttpError(rw, req, err)
			return
		}
	}
}

func nowHandler(rw http.ResponseWriter, req *http.Request) error {
	usr := contextUser(req)
	return pages.NowPage(usr, time.Now()).Render(req.Context(), rw)
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

	isHxRequest := req.Header.Get("HX-Request")
	if isHxRequest == "true" {
		ErrorPartial(e).Render(req.Context(), rw)
	} else {
		usr := contextUser(req)
		ErrorPage(usr, e).Render(req.Context(), rw)
	}
}
