package httpx

const HeaderContentType = "Content-Type"

const MIMEApplicationJSON = "application/json"

const (
	MsgBadRequest           = "Invalid request"
	MsgInternalError        = "Internal server error"
	MsgMethodNotAllowed     = "Method not allowed for this route"
	MsgRequestBodyTooLarge  = "Request body too large"
	MsgRouteNotFound        = "Route not found"
	MsgServiceUnavailable   = "Service unavailable"
	MsgUnsupportedMediaType = "Unsupported media type"
	MsgValidationError      = "Request validation failed"
)

const (
	CodeInternalError        Code = "internal_error"
	CodeInvalidRequest       Code = "invalid_request"
	CodeMethodNotAllowed     Code = "method_not_allowed"
	CodePayloadTooLarge      Code = "payload_too_large"
	CodeRouteNotFound        Code = "route_not_found"
	CodeServiceUnavailable   Code = "service_unavailable"
	CodeUnsupportedMediaType Code = "unsupported_media_type"
	CodeValidationFailed     Code = "validation_failed"
)

const (
	FieldCodeInvalidFormat FieldCode = "invalid_format"
	FieldCodeNonPositive   FieldCode = "non_positive"
	FieldCodeRequired      FieldCode = "required"
	FieldCodeTooLong       FieldCode = "too_long"
)
