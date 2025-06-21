package config

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Loader handles loading configuration from various sources
type Loader struct{}

// NewLoader creates a new configuration loader
func NewLoader() *Loader {
	return &Loader{}
}

// LoadFromFile loads configuration from a file (JSON or YAML)
func (l *Loader) LoadFromFile(path string, config *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Determine format from file extension or content
	if strings.HasSuffix(path, ".json") {
		return json.Unmarshal(data, config)
	}

	// Default to YAML
	return yaml.Unmarshal(data, config)
}

// LoadFromEnv loads configuration from environment variables
func (l *Loader) LoadFromEnv(config *Config) error {
	return l.loadEnvToStruct(reflect.ValueOf(config).Elem())
}

// loadEnvToStruct recursively loads environment variables into struct
func (l *Loader) loadEnvToStruct(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Get env tag
		envTag := fieldType.Tag.Get("env")
		if envTag != "" {
			if err := l.setFieldFromEnv(field, envTag); err != nil {
				return fmt.Errorf("failed to set field %s: %w", fieldType.Name, err)
			}
		}

		// Recursively handle nested structs
		if field.Kind() == reflect.Struct {
			if err := l.loadEnvToStruct(field); err != nil {
				return err
			}
		}

		// Handle maps (for blockchain networks)
		if field.Kind() == reflect.Map && field.Type().Elem().Kind() == reflect.Struct {
			if field.IsNil() {
				field.Set(reflect.MakeMap(field.Type()))
			}
			// Handle map loading if needed
		}
	}

	return nil
}

// setFieldFromEnv sets a field value from environment variable
func (l *Loader) setFieldFromEnv(field reflect.Value, envKey string) error {
	envValue := os.Getenv(envKey)
	if envValue == "" {
		return nil // Environment variable not set
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(envValue)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			// Handle duration
			duration, err := time.ParseDuration(envValue)
			if err != nil {
				return fmt.Errorf("invalid duration format: %w", err)
			}
			field.SetInt(int64(duration))
		} else {
			// Handle regular integers
			intValue, err := strconv.ParseInt(envValue, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid integer format: %w", err)
			}
			field.SetInt(intValue)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(envValue, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid unsigned integer format: %w", err)
		}
		field.SetUint(uintValue)

	case reflect.Bool:
		boolValue, err := strconv.ParseBool(envValue)
		if err != nil {
			return fmt.Errorf("invalid boolean format: %w", err)
		}
		field.SetBool(boolValue)

	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			// Handle string slices (comma-separated)
			values := strings.Split(envValue, ",")
			for i, v := range values {
				values[i] = strings.TrimSpace(v)
			}
			field.Set(reflect.ValueOf(values))
		}

	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}

	return nil
}

// WriteYAML writes configuration to YAML file
func (l *Loader) WriteYAML(path string, config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// WriteJSON writes configuration to JSON file
func (l *Loader) WriteJSON(path string, config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}
