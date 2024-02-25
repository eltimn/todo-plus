package errs

import (
	"io"
	"log/slog"
	"net/http"
)

type HttpError struct {
	StatusCode int

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
			StatusCode: http.StatusBadRequest,

			Code:    "eof",
			Message: "EOF reading HTTP request body",
		}
		// case sql.ErrNoRows:
		// 	return HttpError{
		// 		StatusCode: http.StatusNotFound,

		// 		Code:    "not_found",
		// 		Message: "Page Not Found",
		// 	}
	}

	return HttpError{
		StatusCode: http.StatusInternalServerError,

		Code:    "internal",
		Message: "Internal server error",
	}
}

var NotFoundError = HttpError{
	StatusCode: http.StatusNotFound,
	Code:       "not_found",
	Message:    "Page Not Found",
}

func ErrAttr(err error) slog.Attr {
	return slog.Any("error", err)
}
