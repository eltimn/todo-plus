package routes

import (
	"eltimn/todo-plus/utils"
	"eltimn/todo-plus/web/components"
	"log/slog"
	"net/http"
	"time"
)

// appError is a custom error type for handling application errors in the http pipeline
type appError struct {
	error
	Message string
	Code    int
}

func (e *appError) Error() string {
	return e.Message
}

type appHandler func(http.ResponseWriter, *http.Request) *appError

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		slog.Error(err.Error(), utils.ErrAttr(err))
		// http.Error(w, err.Error(), err.Code)

		w.WriteHeader(err.Code)

		isHxRequest := r.Header.Get("HX-Request")
		if isHxRequest == "true" {
			ErrorPartial(err).Render(r.Context(), w)
		} else {
			ErrorPage(err).Render(r.Context(), w)
		}
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) *appError {
	name := r.PathValue("name")
	slog.Info("Name", slog.String("name", name))
	if name == "tim" {
		return &appError{Message: "I'm sorry, I can't do that", Code: http.StatusForbidden}
	}
	components.Hello(name).Render(r.Context(), w)
	return nil
}

func Routes() {
	http.Handle("GET /hello/{name}", appHandler(helloHandler))

	http.HandleFunc("GET /now", func(w http.ResponseWriter, r *http.Request) {
		components.DisplayTime(time.Now()).Render(r.Context(), w)
	})

	fs := http.FileServer(http.Dir("./web/static/assets"))
	http.Handle("GET /assets/", http.StripPrefix("/assets/", fs))

	todoRoutes()

	// TODO: Add 404 handler
	// http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
	// 	http.Error(w, "Not Found", http.StatusNotFound)
	// })

}
