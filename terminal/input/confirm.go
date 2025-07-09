package input

import (
    "fmt"
    "strings"
)

// ConfirmConfig configures confirmation behavior
type ConfirmConfig struct {
    Message     string
    Default     bool
    YesLabel    string
    NoLabel     string
    HelpText    string
    RequireExplicit bool
}

// Confirm displays a confirmation prompt
func (ir *InputReader) Confirm(config ConfirmConfig) (bool, error) {
    // Set defaults
    if config.YesLabel == "" {
        config.YesLabel = "yes"
    }
    if config.NoLabel == "" {
        config.NoLabel = "no"
    }
    
    for {
        // Build prompt
        var prompt strings.Builder
        prompt.WriteString(fmt.Sprintf("%s%s%s", ColorYellow+ColorBold, config.Message, ColorReset))
        
        if !config.RequireExplicit {
            if config.Default {
                prompt.WriteString(fmt.Sprintf(" %s[Y/n]%s", ColorDim, ColorReset))
            } else {
                prompt.WriteString(fmt.Sprintf(" %s[y/N]%s", ColorDim, ColorReset))
            }
        } else {
            prompt.WriteString(fmt.Sprintf(" %s[%s/%s]%s", ColorDim, config.YesLabel, config.NoLabel, ColorReset))
        }
        
        if config.HelpText != "" {
            prompt.WriteString(fmt.Sprintf("\n%s%s%s", ColorDim, config.HelpText, ColorReset))
        }
        
        prompt.WriteString(": ")
        
        ir.printf("%s", prompt.String())
        
        // Read input
        input, err := ir.readLine()
        if err != nil {
            return false, fmt.Errorf("failed to read input: %w", err)
        }
        
        input = strings.ToLower(strings.TrimSpace(input))
        
        // Handle empty input (use default)
        if input == "" && !config.RequireExplicit {
            return config.Default, nil
        }
        
        // Check for explicit answers
        if config.RequireExplicit {
            if input == strings.ToLower(config.YesLabel) {
                return true, nil
            }
            if input == strings.ToLower(config.NoLabel) {
                return false, nil
            }
            ir.printf("%s❌ Please enter '%s' or '%s'%s\n", ColorRed, config.YesLabel, config.NoLabel, ColorReset)
            continue
        }
        
        // Standard yes/no parsing
        switch input {
        case "y", "yes", "true", "1":
            return true, nil
        case "n", "no", "false", "0":
            return false, nil
        default:
            ir.printf("%s❌ Please enter 'y' or 'n'%s\n", ColorRed, ColorReset)
        }
    }
}

// SimpleConfirm is a convenience function for basic confirmation
func SimpleConfirm(message string) (bool, error) {
    ir := NewInputReader()
    return ir.Confirm(ConfirmConfig{
        Message: message,
        Default: false,
    })
}

// ConfirmWithDefault prompts with a default value
func ConfirmWithDefault(message string, defaultValue bool) (bool, error) {
    ir := NewInputReader()
    return ir.Confirm(ConfirmConfig{
        Message: message,
        Default: defaultValue,
    })
}

// ConfirmDangerous prompts for dangerous operations (requires explicit confirmation)
func ConfirmDangerous(message string) (bool, error) {
    ir := NewInputReader()
    return ir.Confirm(ConfirmConfig{
        Message:         fmt.Sprintf("%s⚠️  %s%s", ColorRed+ColorBold, message, ColorReset),
        RequireExplicit: true,
        YesLabel:        "CONFIRM",
        NoLabel:         "cancel",
        HelpText:        "This action cannot be undone. Type 'CONFIRM' to proceed.",
    })
}

// ConfirmFinancial prompts for financial operations
func ConfirmFinancial(operation, amount string) (bool, error) {
    ir := NewInputReader()
    message := fmt.Sprintf("💰 %s %s", operation, amount)
    return ir.Confirm(ConfirmConfig{
        Message:  message,
        Default:  false,
        HelpText: "Please review the transaction details carefully.",
    })
}

// ConfirmDeployment prompts for deployment operations
func ConfirmDeployment(environment string) (bool, error) {
    ir := NewInputReader()
    
    var message string
    var requireExplicit bool
    
    switch strings.ToLower(environment) {
    case "production", "prod", "mainnet":
        message = fmt.Sprintf("%s🚀 Deploy to PRODUCTION%s", ColorRed+ColorBold, ColorReset)
        requireExplicit = true
    case "staging", "testnet":
        message = fmt.Sprintf("%s🧪 Deploy to STAGING%s", ColorYellow+ColorBold, ColorReset)
        requireExplicit = false
    default:
        message = fmt.Sprintf("%s🔧 Deploy to %s%s", ColorBlue+ColorBold, environment, ColorReset)
        requireExplicit = false
    }
    
    config := ConfirmConfig{
        Message:         message,
        RequireExplicit: requireExplicit,
        HelpText:        "Make sure all tests pass and code is reviewed.",
    }
    
    if requireExplicit {
        config.YesLabel = "DEPLOY"
        config.NoLabel = "cancel"
    }
    
    return ir.Confirm(config)
}

// ConfirmCircuitSetup prompts for ZK circuit setup
func ConfirmCircuitSetup(circuit string) (bool, error) {
    ir := NewInputReader()
    return ir.Confirm(ConfirmConfig{
        Message:         fmt.Sprintf("🔐 Generate setup for circuit '%s'", circuit),
        Default:         false,
        RequireExplicit: true,
        YesLabel:        "GENERATE",
        NoLabel:         "cancel",
        HelpText:        "This will generate new proving and verifying keys. Previous keys will be backed up.",
    })
}

// ConfirmReserveOperation prompts for reserve operations
func ConfirmReserveOperation(operation string, threshold float64) (bool, error) {
    ir := NewInputReader()
    
    thresholdStr := fmt.Sprintf("%.2f%%", threshold)
    if threshold < 100 {
        thresholdStr = fmt.Sprintf("%s%s%s", ColorRed, thresholdStr, ColorReset)
    }
    
    return ir.Confirm(ConfirmConfig{
        Message:  fmt.Sprintf("📊 %s (Reserve ratio: %s)", operation, thresholdStr),
        Default:  false,
        HelpText: "This will affect the stablecoin reserve backing.",
    })
}

// Multi-step confirmation for critical operations
func ConfirmMultiStep(steps []string) (bool, error) {
    ir := NewInputReader()
    
    ir.printf("%s🔐 Multi-step confirmation required%s\n", ColorRed+ColorBold, ColorReset)
    ir.printf("%sThe following operations will be performed:%s\n", ColorYellow, ColorReset)
    
    for i, step := range steps {
        ir.printf("  %s%d.%s %s\n", ColorCyan, i+1, ColorReset, step)
    }
    
    // First confirmation
    confirmed, err := ir.Confirm(ConfirmConfig{
        Message: "Do you want to proceed with these operations?",
        Default: false,
    })
    
    if err != nil || !confirmed {
        return false, err
    }
    
    // Second confirmation for extra safety
    return ir.Confirm(ConfirmConfig{
        Message:         "Are you absolutely sure?",
        RequireExplicit: true,
        YesLabel:        "YES",
        NoLabel:         "NO",
        HelpText:        "Type 'YES' to confirm all operations.",
    })
}

// ConfirmWithCode prompts for confirmation with a verification code
func ConfirmWithCode(message string, expectedCode string) (bool, error) {
    ir := NewInputReader()
    
    ir.printf("%s%s%s\n", ColorYellow+ColorBold, message, ColorReset)
    ir.printf("%sVerification code: %s%s%s\n", ColorDim, ColorBold, expectedCode, ColorReset)
    
    for attempts := 0; attempts < 3; attempts++ {
        ir.printf("Enter the verification code to confirm: ")
        
        input, err := ir.readLine()
        if err != nil {
            return false, fmt.Errorf("failed to read input: %w", err)
        }
        
        if strings.TrimSpace(input) == expectedCode {
            ir.printf("%s✅ Code verified%s\n", ColorGreen, ColorReset)
            return true, nil
        }
        
        remaining := 2 - attempts
        if remaining > 0 {
            ir.printf("%s❌ Incorrect code. %d attempts remaining%s\n", ColorRed, remaining, ColorReset)
        }
    }
    
    ir.printf("%s❌ Too many failed attempts%s\n", ColorRed, ColorReset)
    return false, nil
}