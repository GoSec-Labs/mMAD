package input

import (
	"bufio"
	"fmt"
	"strings"
)

// Prompt displays a text input prompt
func (ir *InputReader) Prompt(config PromptConfig) (string, error) {
	for {
		// Display prompt message
		ir.printf("%s%s%s", ColorCyan+ColorBold, config.Message, ColorReset)

		// Show default value if provided
		if config.Default != "" {
			ir.printf(" %s[%s]%s", ColorDim, config.Default, ColorReset)
		}

		// Show placeholder if provided
		if config.Placeholder != "" {
			ir.printf(" %s(%s)%s", ColorDim, config.Placeholder, ColorReset)
		}

		ir.printf(": ")

		// Read input
		input, err := ir.readLine()
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}

		// Use default if no input provided
		if input == "" && config.Default != "" {
			input = config.Default
		}

		// Check if required
		if config.Required && input == "" {
			ir.printf("%s❌ This field is required%s\n", ColorRed, ColorReset)
			continue
		}

		// Validate input
		if config.Validator != nil {
			if err := config.Validator(input); err != nil {
				ir.printf("%s❌ %s%s\n", ColorRed, err.Error(), ColorReset)
				if config.HelpText != "" {
					ir.printf("%s💡 %s%s\n", ColorYellow, config.HelpText, ColorReset)
				}
				continue
			}
		}

		return input, nil
	}
}

// SimplePrompt is a convenience function for basic prompts
func SimplePrompt(message string) (string, error) {
	ir := NewInputReader()
	return ir.Prompt(PromptConfig{
		Message:  message,
		Required: true,
	})
}

// PromptWithDefault prompts with a default value
func PromptWithDefault(message, defaultValue string) (string, error) {
	ir := NewInputReader()
	return ir.Prompt(PromptConfig{
		Message: message,
		Default: defaultValue,
	})
}

// PromptRequired prompts for required input
func PromptRequired(message string) (string, error) {
	ir := NewInputReader()
	return ir.Prompt(PromptConfig{
		Message:  message,
		Required: true,
	})
}

// Common validators
func ValidateNotEmpty(input string) error {
	if strings.TrimSpace(input) == "" {
		return fmt.Errorf("input cannot be empty")
	}
	return nil
}

func ValidateNumber(input string) error {
	if input == "" {
		return fmt.Errorf("please enter a number")
	}
	for _, char := range input {
		if char < '0' || char > '9' {
			return fmt.Errorf("please enter a valid number")
		}
	}
	return nil
}

func ValidateEmail(input string) error {
	if !strings.Contains(input, "@") || !strings.Contains(input, ".") {
		return fmt.Errorf("please enter a valid email address")
	}
	return nil
}

func ValidateURL(input string) error {
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}
	return nil
}

// Multi-line prompt for longer text
func (ir *InputReader) MultiLinePrompt(message string) (string, error) {
	ir.printf("%s%s%s\n", ColorCyan+ColorBold, message, ColorReset)
	ir.printf("%sPress Ctrl+D when finished, or enter '.' on a new line%s\n", ColorDim, ColorReset)

	var lines []string
	scanner := bufio.NewScanner(ir.reader)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "." {
			break
		}
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read multi-line input: %w", err)
	}

	return strings.Join(lines, "\n"), nil
}
