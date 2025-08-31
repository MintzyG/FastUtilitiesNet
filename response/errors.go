package response

import (
	"errors"
	"fmt"
)

var (
	ErrConfigInvalid     = errors.New("configuration error")
	ErrValidationFailed  = errors.New("validation failed")
	ErrSizeLimitExceeded = errors.New("size limit exceeded")
	ErrEncodingFailed    = errors.New("encoding failed")
	ErrTraceFailed       = errors.New("trace error")
	ErrInterceptorFailed = errors.New("interceptor error")
)

type ConfigError struct {
	Field string
	Msg   string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("invalid configuration: %s - %s", e.Field, e.Msg)
}

type ValidationError struct {
	Field   string
	Message string
	Value   any
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed on '%s': %s (value=%v)", e.Field, e.Message, e.Value)
}

type SizeLimitError struct {
	Size int
	Max  int
}

func (e *SizeLimitError) Error() string {
	return fmt.Sprintf("response size (%d bytes) exceeds limit (%d bytes)", e.Size, e.Max)
}

type EncodingError struct {
	Inner error
}

func (e *EncodingError) Error() string {
	return fmt.Sprintf("encoding failed: %v", e.Inner)
}

type TraceError struct {
	Msg string
}

func (e *TraceError) Error() string {
	return fmt.Sprintf("trace error: %s", e.Msg)
}

type InterceptorLimitError struct {
	Current int
	Max     int
}

func (e *InterceptorLimitError) Error() string {
	return fmt.Sprintf("maximum number of interceptors reached: %d/%d", e.Current, e.Max)
}

type StatusCodeError struct {
	Code int
}

func (e *StatusCodeError) Error() string {
	return fmt.Sprintf("invalid HTTP status code: %d", e.Code)
}
