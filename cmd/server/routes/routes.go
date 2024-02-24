package routes

import (
	"eltimn/todo-plus/cmd/server/router"
	"eltimn/todo-plus/utils"
	"eltimn/todo-plus/web/pages"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func Routes() *router.Router {
	opts := router.RouterOpts{HandleError: handleHttpError}
	rtr := router.NewRouter(&opts, router.Mid(0))

	// serve static files
	fs := http.FileServer(http.Dir("./dist/assets"))
	rtr.ServeMux.Handle("GET /assets/", http.StripPrefix("/assets/", fs))

	todoRoutes(rtr)

	rtr.Get("/hello/{name}", helloHandler)
	rtr.Get("/now", nowHandler)

	// Handles the home page and all non-matches
	rtr.ServeMux.Handle("/", http.HandlerFunc(homeHandler))

	return rtr
}

func helloHandler(w http.ResponseWriter, req *http.Request) error {
	name := req.PathValue("name")
	slog.Debug("Name", slog.String("name", name))
	if name == "tim" {
		return fmt.Errorf("sorry, I can't do that")
	}
	return pages.Hello(name).Render(req.Context(), w)
}

func homeHandler(w http.ResponseWriter, req *http.Request) {
	slog.Debug("URL", slog.String("url", req.URL.Path))
	if req.URL.Path != "/" || req.Method != http.MethodGet {
		handleHttpError(w, req, utils.NotFoundError)
		return
	}

	err := pages.HomePage().Render(req.Context(), w)
	if err != nil {
		handleHttpError(w, req, err)
	}
}

func nowHandler(w http.ResponseWriter, req *http.Request) error {
	return pages.NowPage(time.Now()).Render(req.Context(), w)
}

func handleHttpError(w http.ResponseWriter, req *http.Request, err error) {
	slog.Error(err.Error(), utils.ErrAttr(err))

	// Check if the error was an HttpError or a regular error.
	var e utils.HttpError
	switch err := err.(type) {
	case nil:
		e = utils.NewHttpError(fmt.Errorf("nil error"))
	case utils.HttpError: // already an HttpError
		e = err
	default:
		e = utils.NewHttpError(err)
	}

	w.WriteHeader(e.StatusCode)

	isHxRequest := req.Header.Get("HX-Request")
	if isHxRequest == "true" {
		ErrorPartial(e).Render(req.Context(), w)
	} else {
		ErrorPage(e).Render(req.Context(), w)
	}
}
