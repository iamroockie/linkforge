package httpx

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"mime"
	"net/http"
)

type Map map[string]any

func WriteJSON(w http.ResponseWriter, r *http.Request, status int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		status = http.StatusInternalServerError
		data, _ = json.Marshal(ErrorEnvelope{InternalError(nil)})
		ReportError(r.Context(), fmt.Errorf("marshal response: %w", err))
	}

	w.Header().Set(HeaderContentType, MIMEApplicationJSON)
	w.WriteHeader(status)
	if _, err = w.Write(data); err != nil {
		ReportError(r.Context(), fmt.Errorf("write response: %w", err))
	}
}

func ParseRequestJSON[T any](r *http.Request) (T, error) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get(HeaderContentType))
	if err != nil || mediaType != MIMEApplicationJSON {
		return *new(T), NewError(
			http.StatusUnsupportedMediaType,
			CodeUnsupportedMediaType,
			MsgUnsupportedMediaType,
			err,
		)
	}

	var payload T

	err = json.UnmarshalRead(r.Body, &payload, json.RejectUnknownMembers(true))
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			return *new(T), NewError(
				http.StatusRequestEntityTooLarge,
				CodePayloadTooLarge,
				MsgRequestBodyTooLarge,
				err,
			)
		}

		return *new(T), BadRequestError()
	}

	return payload, nil
}
