package httpx

import (
	"errors"
	"net/http"
)

type Response struct {
	Status  int
	Payload any
}

func NewResponse(status int, payload any) *Response {
	return &Response{Status: status, Payload: payload}
}

type HandlerFunc func(w http.ResponseWriter, r *http.Request) (*Response, error)

func Handle(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := fn(w, r)
		if err != nil {
			e, ok := errors.AsType[Error](err)
			if !ok {
				e = InternalError(err)
			}

			if e.Status >= http.StatusInternalServerError {
				ReportError(r.Context(), e.Cause)
			}

			resp = NewResponse(e.Status, ErrorEnvelope{e})
		}
		if resp == nil {
			return
		}

		WriteJSON(w, r, resp.Status, resp.Payload)
	}
}
