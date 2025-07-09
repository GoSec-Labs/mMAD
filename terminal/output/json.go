package output

import (
	"encoding/json"
	"time"
)

// JSONResponse wraps output with metadata
type JSONResponse struct {
	Success   bool        `json:"success"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
}

type Meta struct {
	Component string `json:"component,omitempty"`
	Version   string `json:"version,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

func (w *Writer) writeJSON(data interface{}) error {
	var output interface{}

	// Wrap in JSONResponse if not already wrapped
	if resp, ok := data.(*JSONResponse); ok {
		output = resp
	} else {
		output = &JSONResponse{
			Success:   true,
			Timestamp: time.Now().UTC(),
			Data:      data,
		}
	}

	var err error
	if w.indent {
		err = json.NewEncoder(w.Writer).Encode(output)
	} else {
		encoder := json.NewEncoder(w.Writer)
		encoder.SetIndent("", "  ")
		err = encoder.Encode(output)
	}

	return err
}

// JSONSuccess creates a successful JSON response
func JSONSuccess(data interface{}) *JSONResponse {
	return &JSONResponse{
		Success:   true,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}
}

// JSONError creates an error JSON response
func JSONError(message string) *JSONResponse {
	return &JSONResponse{
		Success:   false,
		Timestamp: time.Now().UTC(),
		Error:     message,
	}
}

// JSONWithMeta creates a JSON response with metadata
func JSONWithMeta(data interface{}, component, version string) *JSONResponse {
	return &JSONResponse{
		Success:   true,
		Timestamp: time.Now().UTC(),
		Data:      data,
		Meta: &Meta{
			Component: component,
			Version:   version,
		},
	}
}
