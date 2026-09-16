package httpx

import (
	"errors"
	"net/http"
)

type ErrorEnvelope struct {
	Error Error `json:"error"`
}

type Code string

type Error struct {
	Status int   `json:"-"`
	Cause  error `json:"-"`

	Code    Code         `json:"code"`
	Message string       `json:"message"`
	Details []FieldError `json:"details,omitempty"`
}

func (e Error) Error() string {
	msg := string(e.Code) + ": " + e.Message
	if e.Cause != nil {
		msg += ": " + e.Cause.Error()
	}

	return msg
}

func (e Error) Unwrap() error { return e.Cause }

type FieldCode string

type FieldError struct {
	Field string    `json:"field"`
	Code  FieldCode `json:"code"`
}

type FieldRule struct {
	Field    string
	Code     FieldCode
	Sentinel error
}

func NewError(status int, code Code, message string, cause error) Error {
	return Error{Status: status, Code: code, Message: message, Cause: cause}
}

func InternalError(cause error) Error {
	return Error{
		Status:  http.StatusInternalServerError,
		Code:    CodeInternalError,
		Message: MsgInternalError,
		Cause:   cause,
	}
}

func BadRequestError() Error {
	return Error{
		Status:  http.StatusBadRequest,
		Code:    CodeInvalidRequest,
		Message: MsgBadRequest,
		Cause:   nil,
	}
}

func NotFoundError(code Code, message string) Error {
	return Error{
		Status:  http.StatusNotFound,
		Code:    code,
		Message: message,
		Cause:   nil,
	}
}

func MethodNotAllowedError() Error {
	return Error{
		Status:  http.StatusMethodNotAllowed,
		Code:    CodeMethodNotAllowed,
		Message: MsgMethodNotAllowed,
		Cause:   nil,
	}
}

func ValidationError(details []FieldError) Error {
	return Error{
		Status:  http.StatusUnprocessableEntity,
		Code:    CodeValidationFailed,
		Message: MsgValidationError,
		Details: details,
		Cause:   nil,
	}
}

func MatchFieldErrors(err error, rules []FieldRule) []FieldError {
	reported := make(map[string]bool, len(rules))
	fe := make([]FieldError, 0, len(rules))

	for _, e := range rules {
		if reported[e.Field] || !errors.Is(err, e.Sentinel) {
			continue
		}
		reported[e.Field] = true
		fe = append(fe, FieldError{Field: e.Field, Code: e.Code})
	}

	return fe
}
