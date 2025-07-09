package input

import (
	"fmt"
	"strconv"
	"strings"
)

// SelectConfig configures selection behavior
type SelectConfig struct {
	Message     string
	Options     []Option
	Default     int
	MultiSelect bool
	Required    bool
	HelpText    string
}

// Select displays a selection menu
func (ir *InputReader) Select(config SelectConfig) ([]string, error) {
	if len(config.Options) == 0 {
		return nil, fmt.Errorf("no options provided")
	}

	for {
		// Display message
		ir.printf("%s%s%s\n", ColorCyan+ColorBold, config.Message, ColorReset)

		if config.HelpText != "" {
			ir.printf("%s%s%s\n", ColorDim, config.HelpText, ColorReset)
		}

		// Display options
		for i, option := range config.Options {
			prefix := fmt.Sprintf("%s[%d]%s", ColorYellow, i+1, ColorReset)

			if option.Selected {
				prefix = fmt.Sprintf("%s[%d] ✓%s", ColorGreen, i+1, ColorReset)
			}

			ir.printf("  %s %s", prefix, option.Label)

			if option.Description != "" {
				ir.printf(" %s- %s%s", ColorDim, option.Description, ColorReset)
			}

			ir.printf("\n")
		}

		// Show instructions
		if config.MultiSelect {
			ir.printf("\n%sEnter numbers separated by commas (e.g., 1,3,5) or 'done' to finish:%s ",
				ColorDim, ColorReset)
		} else {
			ir.printf("\n%sEnter option number:%s ", ColorDim, ColorReset)
		}

		// Show default
		if config.Default > 0 && config.Default <= len(config.Options) {
			ir.printf("%s[%d]%s ", ColorDim, config.Default, ColorReset)
		}

		// Read input
		input, err := ir.readLine()
		if err != nil {
			return nil, fmt.Errorf("failed to read input: %w", err)
		}

		// Handle default
		if input == "" && config.Default > 0 {
			input = strconv.Itoa(config.Default)
		}

		// Handle multi-select 'done'
		if config.MultiSelect && strings.ToLower(input) == "done" {
			selected := make([]string, 0)
			for _, option := range config.Options {
				if option.Selected {
					selected = append(selected, option.Value)
				}
			}

			if config.Required && len(selected) == 0 {
				ir.printf("%s❌ Please select at least one option%s\n", ColorRed, ColorReset)
				continue
			}

			return selected, nil
		}

		// Parse selections
		selections, err := ir.parseSelections(input, len(config.Options))
		if err != nil {
			ir.printf("%s❌ %s%s\n", ColorRed, err.Error(), ColorReset)
			continue
		}

		if config.MultiSelect {
			// Toggle selections
			for _, sel := range selections {
				config.Options[sel-1].Selected = !config.Options[sel-1].Selected
			}
			ir.printf("%s✓ Selection updated%s\n", ColorGreen, ColorReset)
		} else {
			// Return single selection
			if len(selections) > 1 {
				ir.printf("%s❌ Please select only one option%s\n", ColorRed, ColorReset)
				continue
			}

			selected := config.Options[selections[0]-1]
			return []string{selected.Value}, nil
		}
	}
}

// parseSelections parses comma-separated selections
func (ir *InputReader) parseSelections(input string, maxOptions int) ([]int, error) {
	if strings.TrimSpace(input) == "" {
		return nil, fmt.Errorf("please enter a selection")
	}

	parts := strings.Split(input, ",")
	selections := make([]int, 0)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		num, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("'%s' is not a valid number", part)
		}

		if num < 1 || num > maxOptions {
			return nil, fmt.Errorf("option %d is not valid (1-%d)", num, maxOptions)
		}

		selections = append(selections, num)
	}

	if len(selections) == 0 {
		return nil, fmt.Errorf("no valid selections found")
	}

	return selections, nil
}

// SimpleSelect is a convenience function for basic selection
func SimpleSelect(message string, options []string) (string, error) {
	opts := make([]Option, len(options))
	for i, opt := range options {
		opts[i] = Option{
			Label: opt,
			Value: opt,
		}
	}

	ir := NewInputReader()
	selected, err := ir.Select(SelectConfig{
		Message: message,
		Options: opts,
	})

	if err != nil {
		return "", err
	}

	return selected[0], nil
}

// MultiSelect allows multiple selections
func MultiSelect(message string, options []string) ([]string, error) {
	opts := make([]Option, len(options))
	for i, opt := range options {
		opts[i] = Option{
			Label: opt,
			Value: opt,
		}
	}

	ir := NewInputReader()
	return ir.Select(SelectConfig{
		Message:     message,
		Options:     opts,
		MultiSelect: true,
	})
}
