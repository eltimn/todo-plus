package errs

import (
	"io"
	"log/slog"
	"net/http"
)

// internal error code
type HttpErrorCode string

const InternalErrorCode HttpErrorCode = "internal"
const BadRequestErrorCode HttpErrorCode = "bad_request"
const NotFoundErrorCode HttpErrorCode = "not_found"

type HttpError struct {
	StatusCode int

	Code    HttpErrorCode `json:"code"`
	Message string        `json:"message"`
}

func (e HttpError) Error() string {
	return e.Message
}

func NewHttpError(err error) HttpError {
	switch err {
	case io.EOF:
		return BadRequestError("EOF reading HTTP request body")
		// case sql.ErrNoRows:
		// 	return HttpError{
		// 		StatusCode: http.StatusNotFound,

		// 		Code:    "not_found",
		// 		Message: "Page Not Found",
		// 	}
	}

	return InternalServerError(err.Error())
}

var NotFoundError = HttpError{
	StatusCode: http.StatusNotFound,
	Code:       NotFoundErrorCode,
	Message:    "Page Not Found",
}

func BadRequestError(msg string) HttpError {
	return HttpError{
		StatusCode: http.StatusBadRequest,
		Code:       BadRequestErrorCode,
		Message:    msg,
	}
}

func InternalServerError(msg string) HttpError {
	return HttpError{
		StatusCode: http.StatusInternalServerError,
		Code:       InternalErrorCode,
		Message:    msg,
	}
}

func ErrAttr(err error) slog.Attr {
	return slog.Any("error", err)
}

type UserError string

func (e UserError) Error() string {
	return string(e)
}

var UserNotFoundError UserError = "User not found."
var PasswordMisMatchError UserError = "Passwords must match."
var InvalidCredentialsError UserError = "Login failed."
