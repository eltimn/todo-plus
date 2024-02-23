package routes

import (
	"eltimn/todo-plus/utils"
	"log/slog"
	"net/http"
)

// https://stackoverflow.com/questions/32485021/simplifying-repetitive-error-handling-with-julienschmidt-httprouter
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
