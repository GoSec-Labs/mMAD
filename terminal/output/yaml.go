package output

import (
	"time"

	"gopkg.in/yaml.v3"
)

// YAMLResponse wraps output for YAML format
type YAMLResponse struct {
	Success   bool        `yaml:"success"`
	Timestamp time.Time   `yaml:"timestamp"`
	Data      interface{} `yaml:"data,omitempty"`
	Error     string      `yaml:"error,omitempty"`
}

func (w *Writer) writeYAML(data interface{}) error {
	var output interface{}

	// Wrap in YAMLResponse if not already wrapped
	if resp, ok := data.(*YAMLResponse); ok {
		output = resp
	} else {
		output = &YAMLResponse{
			Success:   true,
			Timestamp: time.Now().UTC(),
			Data:      data,
		}
	}

	encoder := yaml.NewEncoder(w.Writer)
	encoder.SetIndent(2)
	defer encoder.Close()

	return encoder.Encode(output)
}

// YAMLSuccess creates a successful YAML response
func YAMLSuccess(data interface{}) *YAMLResponse {
	return &YAMLResponse{
		Success:   true,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}
}

// YAMLError creates an error YAML response
func YAMLError(message string) *YAMLResponse {
	return &YAMLResponse{
		Success:   false,
		Timestamp: time.Now().UTC(),
		Error:     message,
	}
}
