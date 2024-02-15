package routes

import (
	"eltimn/todo-plus/utils"
	"io"
	"log/slog"
	"net/http"

	"github.com/uptrace/bunrouter"
)

type HttpError struct {
	statusCode int

	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e HttpError) Error() string {
	return e.Message
}

func NewHttpError(err error) HttpError {
	switch err {
	case io.EOF:
		return HttpError{
			statusCode: http.StatusBadRequest,

			Code:    "eof",
			Message: "EOF reading HTTP request body",
		}
		// case sql.ErrNoRows:
		// 	return HttpError{
		// 		statusCode: http.StatusNotFound,

		// 		Code:    "not_found",
		// 		Message: "Page Not Found",
		// 	}
	}

	return HttpError{
		statusCode: http.StatusInternalServerError,

		Code:    "internal",
		Message: "Internal server error",
	}
}

// func apiErrorHandler(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
// 	return func(w http.ResponseWriter, req bunrouter.Request) error {
// 		// Call the next handler on the chain to get the error.
// 		err := next(w, req)

// 		switch err := err.(type) {
// 		case nil:
// 			// no error
// 		case HttpError: // already a HttpError
// 			w.WriteHeader(err.statusCode)
// 			_ = bunrouter.JSON(w, err)
// 		default:
// 			httpErr := NewHttpError(err)
// 			w.WriteHeader(httpErr.statusCode)
// 			_ = bunrouter.JSON(w, httpErr)
// 		}

// 		return err // return the err in case there other middlewares
// 	}
// }

func webErrorHandler(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		// Call the next handler on the chain to get the error.
		err := next(w, req)

		switch err := err.(type) {
		case nil:
			// no error
		case HttpError: // already a HttpError
			renderWebError(w, req, err)
		default:
			renderWebError(w, req, NewHttpError(err))
		}

		return err // return the err in case there other middlewares
	}
}

func renderWebError(w http.ResponseWriter, req bunrouter.Request, httpErr HttpError) {
	slog.Error(httpErr.Error(), utils.ErrAttr(httpErr))

	w.WriteHeader(httpErr.statusCode)

	// if htmx sent this request, return a partial
	isHxRequest := req.Header.Get("HX-Request")
	if isHxRequest == "true" {
		ErrorPartial(httpErr).Render(req.Context(), w)
	} else {
		ErrorPage(httpErr).Render(req.Context(), w)
	}
}

func NotFoundHandler(w http.ResponseWriter, req bunrouter.Request) error {
	err := HttpError{
		statusCode: http.StatusNotFound,

		Code:    "not_found",
		Message: "Page Not Found",
	}

	renderWebError(w, req, err)

	return nil
}
