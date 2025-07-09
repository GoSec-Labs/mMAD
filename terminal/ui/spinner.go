package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// Spinner animations
var (
	SpinnerDots    = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	SpinnerLine    = []string{"|", "/", "-", "\\"}
	SpinnerClock   = []string{"🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚", "🕛"}
	SpinnerMoon    = []string{"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"}
	SpinnerArrow   = []string{"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"}
	SpinnerBounce  = []string{"⠁", "⠂", "⠄", "⠂"}
	SpinnerFinance = []string{"💰", "💎", "💵", "💴", "💶", "💷"}
)

// Spinner represents a loading spinner
type Spinner struct {
	frames  []string
	message string
	delay   time.Duration
	writer  io.Writer
	active  bool
	done    chan bool
	mu      sync.Mutex
	color   string
	prefix  string
	suffix  string
}

// NewSpinner creates a new spinner
func NewSpinner(frames []string, message string) *Spinner {
	return &Spinner{
		frames:  frames,
		message: message,
		delay:   100 * time.Millisecond,
		writer:  os.Stdout,
		done:    make(chan bool),
		color:   currentScheme.Primary,
		prefix:  "",
		suffix:  "",
	}
}

// NewDefaultSpinner creates a spinner with default settings
func NewDefaultSpinner(message string) *Spinner {
	return NewSpinner(SpinnerDots, message)
}

// NewFinanceSpinner creates a finance-themed spinner
func NewFinanceSpinner(message string) *Spinner {
	return NewSpinner(SpinnerFinance, message)
}

// SetDelay sets the animation delay
func (s *Spinner) SetDelay(delay time.Duration) *Spinner {
	s.delay = delay
	return s
}

// SetWriter sets the output writer
func (s *Spinner) SetWriter(writer io.Writer) *Spinner {
	s.writer = writer
	return s
}

// SetColor sets the spinner color
func (s *Spinner) SetColor(color string) *Spinner {
	s.color = color
	return s
}

// SetPrefix sets text before the spinner
func (s *Spinner) SetPrefix(prefix string) *Spinner {
	s.prefix = prefix
	return s
}

// SetSuffix sets text after the message
func (s *Spinner) SetSuffix(suffix string) *Spinner {
	s.suffix = suffix
	return s
}

// Start begins the spinner animation
func (s *Spinner) Start() *Spinner {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.active {
		return s
	}

	s.active = true
	go s.animate()
	return s
}

// Stop stops the spinner animation
func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active {
		return
	}

	s.active = false
	s.done <- true
	s.clearLine()
}

// Success stops the spinner and shows success message
func (s *Spinner) Success(message string) {
	s.Stop()
	if message == "" {
		message = s.message
	}
	fmt.Fprintf(s.writer, "\r%s%s %s%s%s\n",
		s.prefix,
		Success("✅"),
		message,
		s.suffix,
		strings.Repeat(" ", 10)) // Clear any remaining characters
}

// Error stops the spinner and shows error message
func (s *Spinner) Error(message string) {
	s.Stop()
	if message == "" {
		message = s.message
	}
	fmt.Fprintf(s.writer, "\r%s%s %s%s%s\n",
		s.prefix,
		Error("❌"),
		message,
		s.suffix,
		strings.Repeat(" ", 10))
}

// Warning stops the spinner and shows warning message
func (s *Spinner) Warning(message string) {
	s.Stop()
	if message == "" {
		message = s.message
	}
	fmt.Fprintf(s.writer, "\r%s%s %s%s%s\n",
		s.prefix,
		Warning("⚠️"),
		message,
		s.suffix,
		strings.Repeat(" ", 10))
}

// UpdateMessage updates the spinner message
func (s *Spinner) UpdateMessage(message string) {
	s.mu.Lock()
	s.message = message
	s.mu.Unlock()
}

// animate runs the spinner animation
func (s *Spinner) animate() {
	frameIndex := 0

	for {
		select {
		case <-s.done:
			return
		default:
			s.mu.Lock()
			if !s.active {
				s.mu.Unlock()
				return
			}

			frame := s.frames[frameIndex%len(s.frames)]
			coloredFrame := colorize(s.color, frame)

			fmt.Fprintf(s.writer, "\r%s%s %s%s",
				s.prefix,
				coloredFrame,
				s.message,
				s.suffix)

			frameIndex++
			s.mu.Unlock()

			time.Sleep(s.delay)
		}
	}
}

// clearLine clears the current line
func (s *Spinner) clearLine() {
	fmt.Fprintf(s.writer, "\r%s\r", strings.Repeat(" ", 80))
}

// SpinnerManager manages multiple spinners
type SpinnerManager struct {
	spinners []*Spinner
	mu       sync.Mutex
}

// NewSpinnerManager creates a new spinner manager
func NewSpinnerManager() *SpinnerManager {
	return &SpinnerManager{
		spinners: make([]*Spinner, 0),
	}
}

// Add adds a spinner to the manager
func (sm *SpinnerManager) Add(spinner *Spinner) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.spinners = append(sm.spinners, spinner)
}

// StopAll stops all managed spinners
func (sm *SpinnerManager) StopAll() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, spinner := range sm.spinners {
		spinner.Stop()
	}
	sm.spinners = sm.spinners[:0] // Clear slice
}

// Convenience functions for common use cases
func ShowSpinner(message string, fn func() error) error {
	spinner := NewDefaultSpinner(message).Start()
	err := fn()

	if err != nil {
		spinner.Error("Failed")
		return err
	}

	spinner.Success("Complete")
	return nil
}

func ShowFinanceSpinner(message string, fn func() error) error {
	spinner := NewFinanceSpinner(message).Start()
	err := fn()

	if err != nil {
		spinner.Error("Failed")
		return err
	}

	spinner.Success("Complete")
	return nil
}

// WithSpinner executes a function with a spinner
func WithSpinner(message string, frames []string, fn func() error) error {
	spinner := NewSpinner(frames, message).Start()
	defer spinner.Stop()

	err := fn()

	if err != nil {
		spinner.Error("")
		return err
	}

	spinner.Success("")
	return nil
}
