package response

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

var (
	ErrConfigInvalid     = errors.New("configuration error")
	ErrValidationFailed  = errors.New("validation failed")
	ErrSizeLimitExceeded = errors.New("size limit exceeded")
	ErrEncodingFailed    = errors.New("encoding failed")
	ErrTraceFailed       = errors.New("trace error")
	ErrInterceptorFailed = errors.New("interceptor error")
)

type Config struct {
	MaxTraceSize         int
	ResponseSizeLimit    int // in bytes
	MaxInterceptorAmount int
	DefaultContentType   string
	EnableSizeValidation bool
}

// Default configuration values
var defaultConfig = Config{
	MaxTraceSize:         50,
	ResponseSizeLimit:    10 * 1024 * 1024, // 10MB
	MaxInterceptorAmount: 20,
	DefaultContentType:   "application/json",
	EnableSizeValidation: true,
}

// Global configuration (thread-safe)
var (
	globalConfig   Config
	globalConfigMu sync.RWMutex
)

// Initialize with default config
func init() {
	globalConfig = defaultConfig
}

// SetConfig updates the global configuration
func SetConfig(config Config) {
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()

	// Validate config values
	if config.MaxTraceSize <= 0 {
		config.MaxTraceSize = defaultConfig.MaxTraceSize
	}
	if config.ResponseSizeLimit <= 0 {
		config.ResponseSizeLimit = defaultConfig.ResponseSizeLimit
	}
	if config.MaxInterceptorAmount <= 0 {
		config.MaxInterceptorAmount = defaultConfig.MaxInterceptorAmount
	}
	if config.DefaultContentType == "" {
		config.DefaultContentType = defaultConfig.DefaultContentType
	}

	globalConfig = config
}

// GetConfig returns a copy of the current global configuration
func GetConfig() Config {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig
}

// getConfig is a helper to get current config (internal use)
func getConfig() Config {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig
}

func newResponseWithCode(code int) *Response {
	config := getConfig()
	return &Response{
		Code:        code,
		Timestamp:   time.Now(),
		ContentType: config.DefaultContentType,
	}
}

// Standard HTTP response builders
func OK() *Response                  { return newResponseWithCode(http.StatusOK) }
func Created() *Response             { return newResponseWithCode(http.StatusCreated) }
func Accepted() *Response            { return newResponseWithCode(http.StatusAccepted) }
func NoContent() *Response           { return newResponseWithCode(http.StatusNoContent) }
func BadRequest() *Response          { return newResponseWithCode(http.StatusBadRequest) }
func Unauthorized() *Response        { return newResponseWithCode(http.StatusUnauthorized) }
func PaymentRequired() *Response     { return newResponseWithCode(http.StatusPaymentRequired) }
func Forbidden() *Response           { return newResponseWithCode(http.StatusForbidden) }
func NotFound() *Response            { return newResponseWithCode(http.StatusNotFound) }
func MethodNotAllowed() *Response    { return newResponseWithCode(http.StatusMethodNotAllowed) }
func Conflict() *Response            { return newResponseWithCode(http.StatusConflict) }
func UnprocessableEntity() *Response { return newResponseWithCode(http.StatusUnprocessableEntity) }
func TooManyRequests() *Response     { return newResponseWithCode(http.StatusTooManyRequests) }
func InternalServerError() *Response { return newResponseWithCode(http.StatusInternalServerError) }
func NotImplemented() *Response      { return newResponseWithCode(http.StatusNotImplemented) }
func BadGateway() *Response          { return newResponseWithCode(http.StatusBadGateway) }
func ServiceUnavailable() *Response  { return newResponseWithCode(http.StatusServiceUnavailable) }

type ResponseInterceptor interface {
	// Called when context is available
	Intercept(ctx context.Context, response *Response, statusCode int)

	// Called when no context is available
	InterceptSimple(response *Response, statusCode int)
}

// Thread-safe interceptors registry
var (
	interceptors   []ResponseInterceptor
	interceptorsMu sync.RWMutex
)

// Interceptor should only be added during downtimes or application initializtion
func AddInterceptor(interceptor ResponseInterceptor) error {
	interceptorsMu.Lock()
	defer interceptorsMu.Unlock()

	config := getConfig()
	if len(interceptors) >= config.MaxInterceptorAmount {
		return &InterceptorLimitError{
			Current: len(interceptors),
			Max:     config.MaxInterceptorAmount,
		}
	}

	interceptors = append(interceptors, interceptor)
	return nil
}

func RemoveAllInterceptors() {
	interceptorsMu.Lock()
	defer interceptorsMu.Unlock()
	interceptors = nil
}

func GetInterceptors() []ResponseInterceptor {
	interceptorsMu.RLock()
	defer interceptorsMu.RUnlock()
	// Return a copy to prevent external modification
	result := make([]ResponseInterceptor, len(interceptors))
	copy(result, interceptors)
	return result
}

type Response struct {
	Module      string    `json:"module,omitempty"`
	Message     string    `json:"message,omitempty"`
	Data        any       `json:"data,omitempty"`
	Trace       []string  `json:"trace,omitempty"`
	Timestamp   time.Time `json:"timestamp,omitempty"`
	Code        int       `json:"code,omitempty"`
	ContentType string    `json:"-"`
}

func New(message string) *Response {
	config := getConfig()
	return &Response{
		Message:     message,
		Code:        http.StatusOK,
		Timestamp:   time.Now(),
		ContentType: config.DefaultContentType,
	}
}

func NewError(message string, code int) *Response {
	config := getConfig()
	return &Response{
		Message:     message,
		Code:        code,
		Timestamp:   time.Now(),
		ContentType: config.DefaultContentType,
	}
}

func NewSuccess(message string, data any) *Response {
	config := getConfig()
	return &Response{
		Message:     message,
		Data:        data,
		Code:        http.StatusOK,
		Timestamp:   time.Now(),
		ContentType: config.DefaultContentType,
	}
}

type ValidationErr struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

func WithValidationErrors(errors any) *Response {
	var validationErrs []ValidationErr

	switch v := errors.(type) {
	case ValidationErr:
		validationErrs = []ValidationErr{v}
	case []ValidationErr:
		validationErrs = v
	default:
		return InternalServerError().
			WithMessage("Invalid validation errors type").
			appendTraceInternal(fmt.Sprintf("Expected ValidationErr or []ValidationErr, got %T", errors))
	}

	r := BadRequest().WithMessage("Validation failed")

	for _, err := range validationErrs {
		if err.Value != nil {
			r.appendTraceInternal("(" + err.Field + ") " + err.Message + ": " + fmt.Sprintf("%v", err.Value))
		} else {
			r.appendTraceInternal("(" + err.Field + ") " + err.Message)
		}
	}

	return r
}

func (r *Response) WithContentType(ctype string) *Response {
	r.ContentType = ctype
	return r
}

func (r *Response) WithModule(module string) *Response {
	r.Module = module
	return r
}

func (r *Response) WithMessage(message string) *Response {
	r.Message = message
	return r
}

func (r *Response) WithData(data any) *Response {
	r.Data = data
	return r
}

// Takes in strings, errors and Stringers
func (r *Response) AppendTrace(trace ...any) *Response {
	return r.appendTrace(false, trace...)
}

// AppendTraceInternal is for internal use and can override the last trace entry when full
func (r *Response) appendTraceInternal(trace ...any) *Response {
	return r.appendTrace(true, trace...)
}

// Internal trace appending logic
func (r *Response) appendTrace(force bool, trace ...any) *Response {
	config := getConfig()

	for _, t := range trace {
		var traceStr string
		switch v := t.(type) {
		case string:
			traceStr = v
		case error:
			traceStr = v.Error()
		case fmt.Stringer:
			traceStr = v.String()
		default:
			continue
		}

		if len(r.Trace) < config.MaxTraceSize {
			r.Trace = append(r.Trace, traceStr)
		} else {
			if force && config.MaxTraceSize > 0 {
				r.Trace[config.MaxTraceSize-1] = traceStr
			} else if !force && config.MaxTraceSize > 0 {
				truncMsg := fmt.Sprintf("... (trace truncated, max size: %d)", config.MaxTraceSize)
				if r.Trace[config.MaxTraceSize-1] != truncMsg {
					r.Trace[config.MaxTraceSize-1] = truncMsg
				}
				break
			}
		}
	}
	return r
}

// Does nothing unless using a custom response
func (r *Response) WithCode(code int) *Response {
	if err := validateStatusCode(code); err != nil {
		return InternalServerError().
			WithMessage("Invalid status code set").
			appendTraceInternal(err)
	} else {
		r.Code = code
	}
	return r
}

func validateStatusCode(code int) error {
	if code < 100 || code > 599 {
		return &StatusCodeError{
			Code: code,
		}
	}
	return nil
}

type sizeEstimator struct {
	size int
}

func (s *sizeEstimator) Write(p []byte) (n int, err error) {
	s.size += len(p)
	return len(p), nil
}

func (r *Response) estimateSize() (int, error) {
	estimator := &sizeEstimator{}
	encoder := json.NewEncoder(estimator)
	if err := encoder.Encode(r); err != nil {
		return 0, err
	}
	return estimator.size, nil
}

// validateResponseSize checks if the response size is within limits
func (r *Response) validateResponseSize() error {
	config := getConfig()
	if !config.EnableSizeValidation {
		return nil
	}

	size, err := r.estimateSize()
	if err != nil {
		return &EncodingError{Inner: err}
	}

	if size > config.ResponseSizeLimit {
		return &SizeLimitError{
			Size: size,
			Max:  config.ResponseSizeLimit,
		}
	}

	return nil
}

// For when you have context (web servers, etc.)
func (r *Response) SendWithContext(ctx context.Context, w http.ResponseWriter) {
	if err := r.validateResponseSize(); err != nil {
		// Create a new error response that fits within limits
		errorResp := InternalServerError().
			WithMessage("Response too large").
			WithContentType(r.ContentType)
		errorResp.sendInternal(ctx, w)
		return
	}

	r.sendInternal(ctx, w)
}

// For when you don't have context (simple cases, tests, etc.)
func (r *Response) Send(w http.ResponseWriter) {
	r.SendWithContext(context.Background(), w)
}

// Internal send method to avoid code duplication
func (r *Response) sendInternal(ctx context.Context, w http.ResponseWriter) {
	interceptorsMu.RLock()
	currentInterceptors := make([]ResponseInterceptor, len(interceptors))
	copy(currentInterceptors, interceptors)
	interceptorsMu.RUnlock()

	for _, interceptor := range currentInterceptors {
		if ctx != nil && ctx != context.Background() {
			interceptor.Intercept(ctx, r, r.Code)
		} else {
			interceptor.InterceptSimple(r, r.Code)
		}
	}

	w.Header().Set("Content-Type", r.ContentType)
	w.WriteHeader(r.Code)

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(r); err != nil {
		// If encoding fails, we can't send the original response so we leave it to Interceptors
		r.appendTraceInternal((&EncodingError{Inner: err}).Error())
	}
}

func (r *Response) GetResponseStats() map[string]any {
	data, _ := json.Marshal(r)
	return map[string]any{
		"size_bytes":   len(data),
		"trace_count":  len(r.Trace),
		"content_type": r.ContentType,
		"status_code":  r.Code,
	}
}

func (r *Response) IsWithinLimits() bool {
	config := getConfig()

	if len(r.Trace) > config.MaxTraceSize {
		return false
	}

	if config.EnableSizeValidation {
		if err := r.validateResponseSize(); err != nil {
			return false
		}
	}

	return true
}
