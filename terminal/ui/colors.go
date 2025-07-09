package ui

import (
	"fmt"
	"os"
	"strings"
)

// ANSI Color codes
const (
	// Reset
	Reset = "\033[0m"

	// Text styles
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"
	Blink     = "\033[5m"
	Reverse   = "\033[7m"

	// Foreground colors
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	// Bright foreground colors
	BrightBlack   = "\033[90m"
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"

	// Background colors
	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"
)

// ColorScheme defines a set of colors for different UI elements
type ColorScheme struct {
	Primary   string
	Secondary string
	Success   string
	Warning   string
	Error     string
	Info      string
	Muted     string
	Accent    string
}

// Predefined color schemes
var (
	// DefaultScheme - Professional blue theme
	DefaultScheme = ColorScheme{
		Primary:   Cyan + Bold,
		Secondary: Blue,
		Success:   Green + Bold,
		Warning:   Yellow + Bold,
		Error:     Red + Bold,
		Info:      BrightBlue,
		Muted:     Dim,
		Accent:    Magenta,
	}

	// DarkScheme - Dark theme with bright accents
	DarkScheme = ColorScheme{
		Primary:   BrightCyan + Bold,
		Secondary: BrightBlue,
		Success:   BrightGreen + Bold,
		Warning:   BrightYellow + Bold,
		Error:     BrightRed + Bold,
		Info:      BrightWhite,
		Muted:     BrightBlack,
		Accent:    BrightMagenta,
	}

	// MonochromeScheme - No colors, just styles
	MonochromeScheme = ColorScheme{
		Primary:   Bold,
		Secondary: "",
		Success:   Bold,
		Warning:   Bold,
		Error:     Bold + Underline,
		Info:      "",
		Muted:     Dim,
		Accent:    Underline,
	}

	// FinancialScheme - Professional financial theme
	FinancialScheme = ColorScheme{
		Primary:   Blue + Bold,
		Secondary: Cyan,
		Success:   Green + Bold,
		Warning:   Yellow + Bold,
		Error:     Red + Bold,
		Info:      White,
		Muted:     BrightBlack,
		Accent:    Magenta + Bold,
	}
)

// Current active color scheme
var currentScheme = DefaultScheme

// SetColorScheme sets the active color scheme
func SetColorScheme(scheme ColorScheme) {
	currentScheme = scheme
}

// GetColorScheme returns the current color scheme
func GetColorScheme() ColorScheme {
	return currentScheme
}

// DisableColors disables color output
func DisableColors() {
	currentScheme = MonochromeScheme
}

// IsColorEnabled checks if colors are supported
func IsColorEnabled() bool {
	term := os.Getenv("TERM")
	return term != "" && term != "dumb" &&
		os.Getenv("NO_COLOR") == "" &&
		os.Getenv("TERM_PROGRAM") != ""
}

// Color utility functions
func Primary(text string) string {
	return colorize(currentScheme.Primary, text)
}

func Secondary(text string) string {
	return colorize(currentScheme.Secondary, text)
}

func Success(text string) string {
	return colorize(currentScheme.Success, text)
}

func Warning(text string) string {
	return colorize(currentScheme.Warning, text)
}

func Error(text string) string {
	return colorize(currentScheme.Error, text)
}

func Info(text string) string {
	return colorize(currentScheme.Info, text)
}

func Muted(text string) string {
	return colorize(currentScheme.Muted, text)
}

func Accent(text string) string {
	return colorize(currentScheme.Accent, text)
}

// colorize applies color to text
func colorize(color, text string) string {
	if !IsColorEnabled() {
		return text
	}
	return color + text + Reset
}

// Status indicators with emojis and colors
func SuccessIcon(text string) string {
	return Success("✅ " + text)
}

func ErrorIcon(text string) string {
	return Error("❌ " + text)
}

func WarningIcon(text string) string {
	return Warning("⚠️  " + text)
}

func InfoIcon(text string) string {
	return Info("🔵 " + text)
}

func LoadingIcon(text string) string {
	return Primary("⏳ " + text)
}

func CheckIcon(text string) string {
	return Success("✓ " + text)
}

func CrossIcon(text string) string {
	return Error("✗ " + text)
}

// Financial-specific indicators
func ProfitIcon(text string) string {
	return Success("📈 " + text)
}

func LossIcon(text string) string {
	return Error("📉 " + text)
}

func MoneyIcon(text string) string {
	return Primary("💰 " + text)
}

func SecurityIcon(text string) string {
	return Accent("🔐 " + text)
}

func ProofIcon(text string) string {
	return Primary("🔒 " + text)
}

// Header creates a styled header
func Header(title string) string {
	border := strings.Repeat("═", len(title)+4)
	return fmt.Sprintf("\n%s\n%s %s %s\n%s\n",
		Primary(border),
		Primary("║"),
		Primary(Bold+title),
		Primary("║"),
		Primary(border))
}

// Section creates a section divider
func Section(title string) string {
	return fmt.Sprintf("\n%s %s\n%s\n",
		Primary("▶"),
		Primary(Bold+title),
		Primary(strings.Repeat("─", len(title)+2)))
}

// Box creates a bordered box around text
func Box(content string, width int) string {
	lines := strings.Split(content, "\n")
	if width == 0 {
		// Auto-calculate width
		for _, line := range lines {
			if len(line) > width {
				width = len(line)
			}
		}
		width += 4 // padding
	}

	var result strings.Builder

	// Top border
	result.WriteString(Primary("┌" + strings.Repeat("─", width-2) + "┐\n"))

	// Content lines
	for _, line := range lines {
		padding := width - len(line) - 4
		if padding < 0 {
			padding = 0
		}
		result.WriteString(fmt.Sprintf("%s│ %s%s │%s\n",
			Primary(""), line, strings.Repeat(" ", padding), Reset))
	}

	// Bottom border
	result.WriteString(Primary("└" + strings.Repeat("─", width-2) + "┘"))

	return result.String()
}

// Gradient applies a color gradient effect (simplified)
func Gradient(text string, startColor, endColor string) string {
	if !IsColorEnabled() {
		return text
	}
	// For now, just alternate colors - could be enhanced with true gradient
	var result strings.Builder
	for i, char := range text {
		if i%2 == 0 {
			result.WriteString(startColor + string(char) + Reset)
		} else {
			result.WriteString(endColor + string(char) + Reset)
		}
	}
	return result.String()
}
