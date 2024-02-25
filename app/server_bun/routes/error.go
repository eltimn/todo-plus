package routes

import (
	"eltimn/todo-plus/pkg/errs"
	"log/slog"
	"net/http"

	"github.com/uptrace/bunrouter"
)

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
		case errs.HttpError: // already a HttpError
			renderWebError(w, req, err)
		default:
			renderWebError(w, req, errs.NewHttpError(err))
		}

		return err // return the err in case there other middlewares
	}
}

func renderWebError(w http.ResponseWriter, req bunrouter.Request, httpErr errs.HttpError) {
	slog.Error(httpErr.Error(), errs.ErrAttr(httpErr))

	w.WriteHeader(httpErr.StatusCode)

	// if htmx sent this request, return a partial
	isHxRequest := req.Header.Get("HX-Request")
	if isHxRequest == "true" {
		ErrorPartial(httpErr).Render(req.Context(), w)
	} else {
		ErrorPage(httpErr).Render(req.Context(), w)
	}
}

func NotFoundHandler(w http.ResponseWriter, req bunrouter.Request) error {
	err := errs.NotFoundError

	renderWebError(w, req, err)

	return nil
}
