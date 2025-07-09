package output

import (
	"fmt"
	"io"
	"os"
)

type OutputFormat string

const (
	FormatHuman OutputFormat = "human"
	FormatJSON  OutputFormat = "json"
	FormatYAML  OutputFormat = "yaml"
	FormatCSV   OutputFormat = "csv"
	FormatTable OutputFormat = "table"
)

// write handles different output fornat
type Writer struct {
	format OutputFormat
	Writer io.Writer
	indent bool
}

func NewWriter(format OutputFormat) *Writer {
	return &Writer{
		format: format,
		Writer: os.Stdout,
		indent: true,
	}
}

// SetWriter sets the output destination
func (w *Writer) SetWriter(writer io.Writer) *Writer {
	w.Writer = writer
	return w
}

// SetIndent controls JSON/YAML indentation
func (w *Writer) SetIndent(indent bool) *Writer {
	w.indent = indent
	return w
}

func (w *Writer) Write(data interface{}) error {
	switch w.format {
	case FormatJSON:
		return w.writeJSON(data)
	case FormatYAML:
		return w.writeYAML(data)
	// case FormatCSV:
	// 	return w.writeCSV(data)
	// case FormatTable:
	// 	return w.writeTable(data)
	case FormatHuman:
		return w.writeHuman(data)
	default:
		return fmt.Errorf("unsupported output format: %s", w.format)
	}
}
