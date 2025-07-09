package input

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"golang.org/x/term"
)

// PasswordConfig configures password input behavior
type PasswordConfig struct {
	Message        string
	Confirmation   bool
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireDigit   bool
	RequireSpecial bool
	AllowEmpty     bool
	HelpText       string
	MaskChar       string
}

// Password prompts for password input with masking
func (ir *InputReader) Password(config PasswordConfig) (string, error) {
	// Set defaults
	if config.MaskChar == "" {
		config.MaskChar = "*"
	}
	if config.Message == "" {
		config.Message = "Password"
	}

	for {
		// Display prompt
		ir.printf("%s%s%s: ", ColorCyan+ColorBold, config.Message, ColorReset)

		// Read password without echo
		password, err := ir.readPassword()
		if err != nil {
			return "", fmt.Errorf("failed to read password: %w", err)
		}

		// Check if empty and allowed
		if password == "" && config.AllowEmpty {
			return password, nil
		}

		// Validate password
		if err := validatePassword(password, config); err != nil {
			ir.printf("%s❌ %s%s\n", ColorRed, err.Error(), ColorReset)
			if config.HelpText != "" {
				ir.printf("%s💡 %s%s\n", ColorYellow, config.HelpText, ColorReset)
			}
			continue
		}

		// Confirmation if required
		if config.Confirmation {
			ir.printf("%sConfirm password%s: ", ColorCyan+ColorBold, ColorReset)
			confirmation, err := ir.readPassword()
			if err != nil {
				return "", fmt.Errorf("failed to read confirmation: %w", err)
			}

			if password != confirmation {
				ir.printf("%s❌ Passwords do not match%s\n", ColorRed, ColorReset)
				continue
			}
		}

		ir.printf("%s✅ Password accepted%s\n", ColorGreen, ColorReset)
		return password, nil
	}
}

// readPassword reads password input without echo
func (ir *InputReader) readPassword() (string, error) {
	// Check if input is from a terminal
	if file, ok := ir.reader.(*os.File); ok {
		if term.IsTerminal(int(file.Fd())) {
			password, err := term.ReadPassword(int(file.Fd()))
			fmt.Fprintln(ir.writer) // New line after password input
			return string(password), err
		}
	}

	// Fallback for non-terminal input (testing, pipes, etc.)
	return ir.readLine()
}

// validatePassword validates password against requirements
func validatePassword(password string, config PasswordConfig) error {
	if len(password) < config.MinLength {
		return fmt.Errorf("password must be at least %d characters long", config.MinLength)
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if config.RequireUpper && !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if config.RequireLower && !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if config.RequireDigit && !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}

	if config.RequireSpecial && !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// SimplePassword prompts for a basic password
func SimplePassword(message string) (string, error) {
	ir := NewInputReader()
	return ir.Password(PasswordConfig{
		Message:   message,
		MinLength: 1,
	})
}

// SecurePassword prompts for a secure password with validation
func SecurePassword(message string) (string, error) {
	ir := NewInputReader()
	return ir.Password(PasswordConfig{
		Message:        message,
		Confirmation:   true,
		MinLength:      8,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigit:   true,
		RequireSpecial: true,
		HelpText:       "Password must be 8+ chars with uppercase, lowercase, digit, and special character",
	})
}

// WalletPassword prompts for a wallet/private key password
func WalletPassword() (string, error) {
	ir := NewInputReader()
	return ir.Password(PasswordConfig{
		Message:        "🔐 Wallet Password",
		Confirmation:   true,
		MinLength:      12,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigit:   true,
		RequireSpecial: true,
		HelpText:       "Strong password required for wallet security (12+ characters)",
	})
}

// DatabasePassword prompts for database password
func DatabasePassword() (string, error) {
	ir := NewInputReader()
	return ir.Password(PasswordConfig{
		Message:    "🗄️  Database Password",
		MinLength:  6,
		AllowEmpty: true,
		HelpText:   "Leave empty to use environment variable",
	})
}

// APIKeyInput prompts for API key (treated as password)
func APIKeyInput(service string) (string, error) {
	ir := NewInputReader()
	return ir.Password(PasswordConfig{
		Message:    fmt.Sprintf("🔑 %s API Key", service),
		MinLength:  1,
		AllowEmpty: false,
	})
}

// PrivateKeyInput prompts for private key input
func PrivateKeyInput(keyType string) (string, error) {
	ir := NewInputReader()
	return ir.Password(PasswordConfig{
		Message:   fmt.Sprintf("🔐 %s Private Key", keyType),
		MinLength: 64, // Typical private key length
		HelpText:  "Enter the private key (input will be hidden)",
	})
}

// MnemonicInput prompts for seed phrase/mnemonic
func MnemonicInput() (string, error) {
	ir := NewInputReader()

	ir.printf("%s🔐 Seed Phrase / Mnemonic%s\n", ColorCyan+ColorBold, ColorReset)
	ir.printf("%sEnter your 12 or 24 word seed phrase:%s\n", ColorDim, ColorReset)

	for {
		ir.printf("Mnemonic: ")

		// Read mnemonic (don't mask since it's words, not a password)
		mnemonic, err := ir.readLine()
		if err != nil {
			return "", fmt.Errorf("failed to read mnemonic: %w", err)
		}

		mnemonic = strings.TrimSpace(mnemonic)
		words := strings.Fields(mnemonic)

		if len(words) != 12 && len(words) != 24 {
			ir.printf("%s❌ Seed phrase must be 12 or 24 words%s\n", ColorRed, ColorReset)
			continue
		}

		// Basic validation - all words should be alphabetic
		valid := true
		for _, word := range words {
			if !isAlphabetic(word) {
				valid = false
				break
			}
		}

		if !valid {
			ir.printf("%s❌ Seed phrase contains invalid characters%s\n", ColorRed, ColorReset)
			continue
		}

		ir.printf("%s✅ Seed phrase accepted (%d words)%s\n", ColorGreen, len(words), ColorReset)
		return mnemonic, nil
	}
}

// isAlphabetic checks if string contains only alphabetic characters
func isAlphabetic(s string) bool {
	for _, char := range s {
		if !unicode.IsLetter(char) {
			return false
		}
	}
	return true
}

// PasswordStrength calculates password strength score (0-100)
func PasswordStrength(password string) int {
	if password == "" {
		return 0
	}

	score := 0
	length := len(password)

	// Length scoring
	if length >= 8 {
		score += 25
	} else if length >= 6 {
		score += 15
	} else if length >= 4 {
		score += 5
	}

	// Character variety
	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	varietyCount := 0
	if hasUpper {
		varietyCount++
	}
	if hasLower {
		varietyCount++
	}
	if hasDigit {
		varietyCount++
	}
	if hasSpecial {
		varietyCount++
	}

	score += varietyCount * 15

	// Length bonus
	if length > 12 {
		score += 10
	}

	// Penalty for common patterns
	lower := strings.ToLower(password)
	if strings.Contains(lower, "password") ||
		strings.Contains(lower, "123456") ||
		strings.Contains(lower, "qwerty") {
		score -= 20
	}

	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	return score
}

// PasswordStrengthText returns password strength as text
// func PasswordStrengthText(password string) string {
// 	strength := PasswordStrength(password)

// 	switch {
// 	case strength >= 80:
// 		return colorize(ColorGreen+ColorBold, "Very Strong")
// 	case strength >= 60:
// 		return colorize(ColorGreen, "Strong")
// 	case strength >= 40:
// 		return colorize(ColorYellow, "Medium")
// 	case strength >= 20:
// 		return colorize(ColorRed, "Weak")
// 	default:
// 		return colorize(ColorRed+ColorBold, "Very Weak")
// 	}
// }

// // ShowPasswordStrength displays password strength meter
// func ShowPasswordStrength(password string) {
// 	strength := PasswordStrength(password)
// 	text := PasswordStrengthText(password)

// 	// Progress bar for strength
// 	barWidth := 20
// 	filled := int(float64(barWidth) * float64(strength) / 100)

// 	var bar strings.Builder
// 	bar.WriteString("[")

// 	for i := 0; i < barWidth; i++ {
// 		if i < filled {
// 			if strength >= 60 {
// 				bar.WriteString(colorize(ColorGreen, "█"))
// 			} else if strength >= 40 {
// 				bar.WriteString(colorize(ColorYellow, "█"))
// 			} else {
// 				bar.WriteString(colorize(ColorRed, "█"))
// 			}
// 		} else {
// 			bar.WriteString("░")
// 		}
// 	}

// 	bar.WriteString("]")

// 	fmt.Printf("Password Strength: %s %s (%d/100)\n", bar.String(), text, strength)
// }
