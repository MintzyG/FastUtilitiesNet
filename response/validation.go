package response

import "fmt"

type ValidationTrace struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

func AddValidationErrors(errs ...ValidationTrace) *Response {
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
