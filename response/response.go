package response

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
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

func newBaseResponse(code int, msg ...string) *Response {
	var message string
	if len(msg) > 0 {
		message = msg[0]
	} else {
		message = ""
	}

	config := getConfig()
	return &Response{
		Code:        code,
		Message:     message,
		Timestamp:   time.Now(),
		ContentType: config.DefaultContentType,
	}
}

func Base(cfg ...*Config) *Response {
	var conf *Config
	if len(cfg) > 0 && cfg[0] != nil {
		conf = cfg[0]
	} else {
		c := getConfig()
		conf = &c
	}

	return &Response{
		ContentType: conf.DefaultContentType,
		config:      *conf,
	}
}

// Standard HTTP response builders
func OK(msg ...string) *Response {
	return newBaseResponse(http.StatusOK, msg...)
}
func Created(msg ...string) *Response {
	return newBaseResponse(http.StatusCreated, msg...)
}
func Accepted(msg ...string) *Response {
	return newBaseResponse(http.StatusAccepted, msg...)
}
func NoContent(msg ...string) *Response {
	return newBaseResponse(http.StatusNoContent, msg...)
}
func BadRequest(msg ...string) *Response {
	return newBaseResponse(http.StatusBadRequest, msg...)
}
func Unauthorized(msg ...string) *Response {
	return newBaseResponse(http.StatusUnauthorized, msg...)
}
func PaymentRequired(msg ...string) *Response {
	return newBaseResponse(http.StatusPaymentRequired, msg...)
}
func Forbidden(msg ...string) *Response {
	return newBaseResponse(http.StatusForbidden, msg...)
}
func NotFound(msg ...string) *Response {
	return newBaseResponse(http.StatusNotFound, msg...)
}
func MethodNotAllowed(msg ...string) *Response {
	return newBaseResponse(http.StatusMethodNotAllowed, msg...)
}
func Conflict(msg ...string) *Response {
	return newBaseResponse(http.StatusConflict, msg...)
}
func UnprocessableEntity(msg ...string) *Response {
	return newBaseResponse(http.StatusUnprocessableEntity, msg...)
}
func TooManyRequests(msg ...string) *Response {
	return newBaseResponse(http.StatusTooManyRequests, msg...)
}
func InternalServerError(msg ...string) *Response {
	return newBaseResponse(http.StatusInternalServerError, msg...)
}
func NotImplemented(msg ...string) *Response {
	return newBaseResponse(http.StatusNotImplemented, msg...)
}
func BadGateway(msg ...string) *Response {
	return newBaseResponse(http.StatusBadGateway, msg...)
}
func ServiceUnavailable(msg ...string) *Response {
	return newBaseResponse(http.StatusServiceUnavailable, msg...)
}

func (r *Response) applyMessage(msg ...string) *Response {
	if len(msg) > 0 {
		r.Message = msg[0]
	} else {
		r.Message = ""
	}
	return r
}

func (r *Response) OK(msg ...string) *Response {
	r.Code = http.StatusOK
	r.applyMessage(msg...)
	return r
}
func (r *Response) Created(msg ...string) *Response {
	r.Code = http.StatusCreated
	r.applyMessage(msg...)
	return r
}
func (r *Response) Accepted(msg ...string) *Response {
	r.Code = http.StatusAccepted
	r.applyMessage(msg...)
	return r
}
func (r *Response) NoContent(msg ...string) *Response {
	r.Code = http.StatusNoContent
	r.applyMessage(msg...)
	return r
}
func (r *Response) BadRequest(msg ...string) *Response {
	r.Code = http.StatusBadRequest
	r.applyMessage(msg...)
	return r
}
func (r *Response) Unauthorized(msg ...string) *Response {
	r.Code = http.StatusUnauthorized
	r.applyMessage(msg...)
	return r
}
func (r *Response) PaymentRequired(msg ...string) *Response {
	r.Code = http.StatusPaymentRequired
	r.applyMessage(msg...)
	return r
}
func (r *Response) Forbidden(msg ...string) *Response {
	r.Code = http.StatusForbidden
	r.applyMessage(msg...)
	return r
}
func (r *Response) NotFound(msg ...string) *Response {
	r.Code = http.StatusNotFound
	r.applyMessage(msg...)
	return r
}
func (r *Response) MethodNotAllowed(msg ...string) *Response {
	r.Code = http.StatusMethodNotAllowed
	r.applyMessage(msg...)
	return r
}
func (r *Response) Conflict(msg ...string) *Response {
	r.Code = http.StatusConflict
	r.applyMessage(msg...)
	return r
}
func (r *Response) UnprocessableEntity(msg ...string) *Response {
	r.Code = http.StatusUnprocessableEntity
	r.applyMessage(msg...)
	return r
}
func (r *Response) TooManyRequests(msg ...string) *Response {
	r.Code = http.StatusTooManyRequests
	r.applyMessage(msg...)
	return r
}
func (r *Response) InternalServerError(msg ...string) *Response {
	r.Code = http.StatusInternalServerError
	r.applyMessage(msg...)
	return r
}
func (r *Response) NotImplemented(msg ...string) *Response {
	r.Code = http.StatusNotImplemented
	r.applyMessage(msg...)
	return r
}
func (r *Response) BadGateway(msg ...string) *Response {
	r.Code = http.StatusBadGateway
	r.applyMessage(msg...)
	return r
}
func (r *Response) ServiceUnavailable(msg ...string) *Response {
	r.Code = http.StatusServiceUnavailable
	r.applyMessage(msg...)
	return r
}

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
	Module         string         `json:"module,omitempty"`
	Message        string         `json:"message,omitempty"`
	Data           any            `json:"data,omitempty"`
	Trace          []string       `json:"trace,omitempty"`
	Timestamp      time.Time      `json:"timestamp,omitempty"`
	PaginationData PaginationMeta `json:"pagination,omitempty"`
	Code           int            `json:"code,omitempty"`
	ContentType    string         `json:"-"`
	TracePrefix    string         `json:"-"`
	config         Config         `json:"-"`
}

type ValidationErr struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

func AddValidationErrors(errs ...ValidationErr) *Response {
	if len(errs) == 0 {
		return BadRequest("Validation failed")
	}

	r := BadRequest("Validation failed")

	for _, err := range errs {
		var traceMsg string
		if err.Value != nil {
			traceMsg = fmt.Sprintf("(%s) %s: %v", err.Field, err.Message, err.Value)
		} else {
			traceMsg = fmt.Sprintf("(%s) %s", err.Field, err.Message)
		}
		r.appendTraceInternal("validation", traceMsg)
	}

	return r
}

// WithConfig sets a custom configuration for this specific response instance
// This overrides the global configuration for this response only
func (r *Response) WithConfig(config Config) *Response {
	// Validate and set defaults for invalid config values
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

	r.config = config

	// Update ContentType if it wasn't explicitly set
	if r.ContentType == "" || r.ContentType == getConfig().DefaultContentType {
		r.ContentType = config.DefaultContentType
	}

	return r
}

// getResponseConfig returns the config for this specific response
// Falls back to global config if no specific config is set
func (r *Response) getResponseConfig() Config {
	// Check if this response has a specific config set
	// We detect this by checking if any field differs from zero value
	if r.config.MaxTraceSize > 0 || r.config.ResponseSizeLimit > 0 ||
		r.config.MaxInterceptorAmount > 0 || r.config.DefaultContentType != "" {
		return r.config
	}
	return getConfig()
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
	NextPage   *int  `json:"next_page,omitempty"`
	PrevPage   *int  `json:"prev_page,omitempty"`
}

type PaginationParams struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

const (
	defaultPage  = 1
	defaultLimit = 20
	maxLimit     = 100
)

func ParsePaginationFromQuery(values url.Values) PaginationParams {
	page, err := strconv.Atoi(values.Get("page"))
	if err != nil || page < 1 {
		page = defaultPage
	}

	limit, err := strconv.Atoi(values.Get("limit"))
	if err != nil || limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	return PaginationParams{
		Page:  page,
		Limit: limit,
	}
}

func CreatePaginationMeta(params PaginationParams, total int64) PaginationMeta {
	totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))
	hasNext := params.Page < totalPages
	hasPrev := params.Page > 1

	meta := PaginationMeta{
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}

	if hasNext {
		nextPage := params.Page + 1
		meta.NextPage = &nextPage
	}

	if hasPrev {
		prevPage := params.Page - 1
		meta.PrevPage = &prevPage
	}

	return meta
}

func (r *Response) WithPagination(params PaginationParams, total int64) *Response {
	meta := CreatePaginationMeta(params, total)
	r.PaginationData = meta
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

func (r *Response) WithMsg(message string) *Response {
	r.Message = message
	return r
}

func (r *Response) WithData(data any) *Response {
	r.Data = data
	return r
}

func (r *Response) WithTracePrefix(prefix string) *Response {
	r.TracePrefix = prefix
	return r
}

// Takes in strings, errors and Stringers
func (r *Response) AddTrace(trace ...any) *Response {
	if r.TracePrefix == "" {
		return r.appendTrace("trace", false, trace...)
	}
	return r.appendTrace(r.TracePrefix, false, trace...)
}

// Takes in strings, errors and Stringers
func (r *Response) AddPrefixedTrace(prefix string, trace ...any) *Response {
	if prefix == "" {
		return r.appendTrace("trace", false, trace...)
	}
	return r.appendTrace(prefix, false, trace...)
}

// AppendTraceInternal is for internal use and can override the last trace entry when full
func (r *Response) appendTraceInternal(prefix string, trace ...any) *Response {
	return r.appendTrace(prefix, true, trace...)
}

// Internal trace appending logic
func (r *Response) appendTrace(prefix string, force bool, trace ...any) *Response {
	config := r.getResponseConfig()

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

		traceStrFull := prefix + ": " + traceStr

		if len(r.Trace) < config.MaxTraceSize {
			r.Trace = append(r.Trace, traceStrFull)
		} else {
			if force && config.MaxTraceSize > 0 {
				r.Trace[config.MaxTraceSize-1] = traceStr
			} else if !force && config.MaxTraceSize > 0 {
				truncMsg := fmt.Sprintf("Error: (trace truncated, max size: %d)", config.MaxTraceSize)
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
		return InternalServerError("Invalid status code set").
			appendTraceInternal("error", err)
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
	config := r.getResponseConfig()
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
		errorResp := r.WithCode(http.StatusInternalServerError).WithContentType(getConfig().DefaultContentType)
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
		r.appendTraceInternal("internal error", (&EncodingError{Inner: err}).Error())
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
	config := r.getResponseConfig()

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
