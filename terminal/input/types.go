package input

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Colors for input prompts
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
)

// InputReader handles input operations
type InputReader struct {
	reader io.Reader
	writer io.Writer
}

// NewInputReader creates a new input reader
func NewInputReader() *InputReader {
	return &InputReader{
		reader: os.Stdin,
		writer: os.Stdout,
	}
}

// SetReader sets the input source
func (ir *InputReader) SetReader(reader io.Reader) *InputReader {
	ir.reader = reader
	return ir
}

// SetWriter sets the output destination
func (ir *InputReader) SetWriter(writer io.Writer) *InputReader {
	ir.writer = writer
	return ir
}

// readLine reads a line from input
func (ir *InputReader) readLine() (string, error) {
	scanner := bufio.NewScanner(ir.reader)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()), nil
	}
	return "", scanner.Err()
}

// printf formats and writes to output
func (ir *InputReader) printf(format string, args ...interface{}) {
	fmt.Fprintf(ir.writer, format, args...)
}

// ValidationFunc validates user input
type ValidationFunc func(string) error

// Option represents a selectable option
type Option struct {
	Label       string
	Value       string
	Description string
	Selected    bool
}

// PromptConfig configures prompt behavior
type PromptConfig struct {
	Message     string
	Default     string
	Required    bool
	Validator   ValidationFunc
	HelpText    string
	Placeholder string
}
