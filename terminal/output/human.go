package output

import (
	"fmt"
	"reflect"
	"strings"
)

// Colors for human output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
)

func (w *Writer) writeHuman(data interface{}) error {
	// Handle different data types
	switch v := data.(type) {
	case *JSONResponse:
		return w.writeHumanResponse(v.Success, v.Data, v.Error)
	case *YAMLResponse:
		return w.writeHumanResponse(v.Success, v.Data, v.Error)
	case error:
		return w.writeHumanError(v.Error())
	case string:
		return w.writeHumanString(v)
	default:
		return w.writeHumanStruct(v)
	}
}

func (w *Writer) writeHumanResponse(success bool, data interface{}, errMsg string) error {
	if !success {
		return w.writeHumanError(errMsg)
	}
	return w.writeHumanStruct(data)
}

func (w *Writer) writeHumanError(message string) error {
	_, err := fmt.Fprintf(w.Writer, "%s❌ Error:%s %s\n",
		ColorRed+ColorBold, ColorReset, message)
	return err
}

func (w *Writer) writeHumanString(message string) error {
	_, err := fmt.Fprintf(w.Writer, "%s\n", message)
	return err
}

func (w *Writer) writeHumanStruct(data interface{}) error {
	if data == nil {
		_, err := fmt.Fprintf(w.Writer, "%sNo data%s\n", ColorDim, ColorReset)
		return err
	}

	// Use reflection to pretty print structs
	v := reflect.ValueOf(data)
	t := reflect.TypeOf(data)

	// Handle pointers
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		return w.writeStructFields(v, t)
	case reflect.Slice, reflect.Array:
		return w.writeSlice(v)
	case reflect.Map:
		return w.writeMap(v)
	default:
		_, err := fmt.Fprintf(w.Writer, "%v\n", data)
		return err
	}
}

func (w *Writer) writeStructFields(v reflect.Value, t reflect.Type) error {
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		fieldName := field.Name
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			if idx := strings.Index(tag, ","); idx != -1 {
				fieldName = tag[:idx]
			} else {
				fieldName = tag
			}
		}

		_, err := fmt.Fprintf(w.Writer, "%s%s:%s %v\n",
			ColorCyan+ColorBold, fieldName, ColorReset, value.Interface())
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) writeSlice(v reflect.Value) error {
	for i := 0; i < v.Len(); i++ {
		_, err := fmt.Fprintf(w.Writer, "%s[%d]%s %v\n",
			ColorYellow, i, ColorReset, v.Index(i).Interface())
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) writeMap(v reflect.Value) error {
	for _, key := range v.MapKeys() {
		value := v.MapIndex(key)
		_, err := fmt.Fprintf(w.Writer, "%s%v:%s %v\n",
			ColorCyan+ColorBold, key.Interface(), ColorReset, value.Interface())
		if err != nil {
			return err
		}
	}
	return nil
}

// Utility functions for colored output
func Success(message string) string {
	return fmt.Sprintf("%s✅ %s%s", ColorGreen+ColorBold, message, ColorReset)
}

func Error(message string) string {
	return fmt.Sprintf("%s❌ %s%s", ColorRed+ColorBold, message, ColorReset)
}

func Warning(message string) string {
	return fmt.Sprintf("%s⚠️  %s%s", ColorYellow+ColorBold, message, ColorReset)
}

func Info(message string) string {
	return fmt.Sprintf("%s🔵 %s%s", ColorBlue+ColorBold, message, ColorReset)
}
