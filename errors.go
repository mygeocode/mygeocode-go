package mygeocode

import (
	"encoding/json"
	"fmt"
	"time"
)

// MyGeocodeError is implemented by every error this package returns for a failed call: *ApiError,
// *RequestTimeoutError, *NetworkError and *DecodeError. Use errors.As to reach the concrete type.
type MyGeocodeError interface {
	error
	myGeocodeError()
}

// ApiError is an error the API answered: any HTTP status of 400 or more, or an answer whose status is "error".
type ApiError struct {
	// Status is the HTTP status code.
	Status int
	// Code is the API's error code, such as invalid_request or quota_exceeded; "http_error" when the answer had none.
	Code string
	// Message is the API's explanation, for logs; "HTTP <status>" when the answer had none. Its wording may change,
	// the code does not.
	Message string
	// Param names the parameter at fault, or is empty.
	Param string
	// RequestID is the X-Request-Id header, or empty.
	RequestID string
	// RetryAfter is the wait the API asked for: Retry-After, or on a 429 the time until X-Quota-Reset. Nil when it
	// asked for none.
	RetryAfter *time.Duration
	// Body is the error answer as the API sent it, when it was a JSON object; nil otherwise.
	Body json.RawMessage
}

func (e *ApiError) myGeocodeError() {}

func (e *ApiError) Error() string {
	return fmt.Sprintf("mygeocode: %s: %s (HTTP %d)", e.Code, e.Message, e.Status)
}

// NetworkError reports that no answer arrived: DNS, a refused or reset connection, TLS, or a context the caller
// cancelled (errors.Is(err, context.Canceled) then holds).
type NetworkError struct {
	Err error
}

func (e *NetworkError) myGeocodeError() {}
func (e *NetworkError) Error() string   { return "mygeocode: network error: " + e.Err.Error() }

// Unwrap returns the underlying error.
func (e *NetworkError) Unwrap() error { return e.Err }

// RequestTimeoutError reports that an attempt did not finish within the client's timeout.
type RequestTimeoutError struct {
	Err error
}

func (e *RequestTimeoutError) myGeocodeError() {}
func (e *RequestTimeoutError) Error() string   { return "mygeocode: request timed out: " + e.Err.Error() }

// Unwrap returns the underlying error.
func (e *RequestTimeoutError) Unwrap() error { return e.Err }

// Timeout reports true, like the timeout errors of the net package.
func (e *RequestTimeoutError) Timeout() bool { return true }

// DecodeError reports an answer that is not a JSON object.
type DecodeError struct {
	// Status is the HTTP status code of the answer, or 0 when the request could not be encoded.
	Status int
	// RequestID is the X-Request-Id header, or empty.
	RequestID string
	// Err says what was wrong.
	Err error
}

func (e *DecodeError) myGeocodeError() {}
func (e *DecodeError) Error() string {
	return fmt.Sprintf("mygeocode: cannot read the answer (HTTP %d): %s", e.Status, e.Err)
}

// Unwrap returns the underlying error.
func (e *DecodeError) Unwrap() error { return e.Err }
