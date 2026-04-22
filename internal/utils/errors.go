package utils

type AppError struct {
	Code       string
	Message    string
	StatusCode int
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

type APIError struct {
	Code    int
	Message string
	Type    string
	Details interface{}
}

func (e *APIError) JSON() map[string]interface{} {
	return map[string]interface{}{
		"error": map[string]interface{}{
			"message": e.Message,
			"type":    e.Type,
			"code":    e.Code,
			"details": e.Details,
		},
	}
}

var (
	ErrUnauthorized = &APIError{Code: 401, Message: "Unauthorized", Type: "auth_error", Details: nil}
	ErrForbidden    = &APIError{Code: 403, Message: "Forbidden", Type: "permission_error", Details: nil}
	ErrNotFound     = &APIError{Code: 404, Message: "Not Found", Type: "not_found_error", Details: nil}
	ErrRateLimit    = &APIError{Code: 429, Message: "Rate limit exceeded", Type: "rate_limit_error", Details: nil}
	ErrServerError  = &APIError{Code: 500, Message: "Internal server error", Type: "server_error", Details: nil}
)
