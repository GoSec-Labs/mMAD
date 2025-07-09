package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// ProgressBar represents a progress bar
type ProgressBar struct {
	total       int64
	current     int64
	width       int
	message     string
	writer      io.Writer
	startTime   time.Time
	mu          sync.Mutex
	showPercent bool
	showETA     bool
	showRate    bool
	unit        string
	fillChar    string
	emptyChar   string
	color       string
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int64, message string) *ProgressBar {
	return &ProgressBar{
		total:       total,
		current:     0,
		width:       50,
		message:     message,
		writer:      os.Stdout,
		startTime:   time.Now(),
		showPercent: true,
		showETA:     true,
		showRate:    false,
		unit:        "",
		fillChar:    "█",
		emptyChar:   "░",
		color:       currentScheme.Primary,
	}
}

// SetWidth sets the progress bar width
func (pb *ProgressBar) SetWidth(width int) *ProgressBar {
	pb.width = width
	return pb
}

// SetWriter sets the output writer
func (pb *ProgressBar) SetWriter(writer io.Writer) *ProgressBar {
	pb.writer = writer
	return pb
}

// SetShowPercent toggles percentage display
func (pb *ProgressBar) SetShowPercent(show bool) *ProgressBar {
	pb.showPercent = show
	return pb
}

// SetShowETA toggles ETA display
func (pb *ProgressBar) SetShowETA(show bool) *ProgressBar {
	pb.showETA = show
	return pb
}

// SetShowRate toggles rate display
func (pb *ProgressBar) SetShowRate(show bool, unit string) *ProgressBar {
	pb.showRate = show
	pb.unit = unit
	return pb
}

// SetChars sets the fill and empty characters
func (pb *ProgressBar) SetChars(fill, empty string) *ProgressBar {
	pb.fillChar = fill
	pb.emptyChar = empty
	return pb
}

// SetColor sets the progress bar color
func (pb *ProgressBar) SetColor(color string) *ProgressBar {
	pb.color = color
	return pb
}

// Update updates the progress
func (pb *ProgressBar) Update(current int64) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.current = current
	if pb.current > pb.total {
		pb.current = pb.total
	}

	pb.render()
}

// Increment increments progress by 1
func (pb *ProgressBar) Increment() {
	pb.Update(pb.current + 1)
}

// Add adds to the current progress
func (pb *ProgressBar) Add(delta int64) {
	pb.Update(pb.current + delta)
}

// Finish completes the progress bar
func (pb *ProgressBar) Finish() {
	pb.Update(pb.total)
	fmt.Fprintln(pb.writer) // New line after completion
}

// render draws the progress bar
func (pb *ProgressBar) render() {
	percent := float64(pb.current) / float64(pb.total) * 100
	filled := int(float64(pb.width) * float64(pb.current) / float64(pb.total))

	// Build progress bar
	var bar strings.Builder
	bar.WriteString(colorize(pb.color, strings.Repeat(pb.fillChar, filled)))
	bar.WriteString(strings.Repeat(pb.emptyChar, pb.width-filled))

	// Build status line
	var status strings.Builder

	// Message
	if pb.message != "" {
		status.WriteString(fmt.Sprintf("%s ", pb.message))
	}

	// Progress bar
	status.WriteString(fmt.Sprintf("[%s] ", bar.String()))

	// Current/Total
	status.WriteString(fmt.Sprintf("%s/%s", formatNumber(pb.current), formatNumber(pb.total)))

	// Percentage
	if pb.showPercent {
		status.WriteString(fmt.Sprintf(" %.1f%%", percent))
	}

	// ETA
	if pb.showETA && pb.current > 0 {
		elapsed := time.Since(pb.startTime)
		rate := float64(pb.current) / elapsed.Seconds()
		if rate > 0 {
			remaining := time.Duration(float64(pb.total-pb.current)/rate) * time.Second
			status.WriteString(fmt.Sprintf(" ETA: %s", formatDuration(remaining)))
		}
	}

	// Rate
	if pb.showRate && pb.current > 0 {
		elapsed := time.Since(pb.startTime)
		rate := float64(pb.current) / elapsed.Seconds()
		status.WriteString(fmt.Sprintf(" %.1f %s/s", rate, pb.unit))
	}

	// Output with carriage return
	fmt.Fprintf(pb.writer, "\r%s", status.String())
}

// MultiProgressBar manages multiple progress bars
type MultiProgressBar struct {
	bars   []*ProgressBar
	writer io.Writer
	mu     sync.Mutex
}

// NewMultiProgressBar creates a new multi-progress bar
func NewMultiProgressBar() *MultiProgressBar {
	return &MultiProgressBar{
		bars:   make([]*ProgressBar, 0),
		writer: os.Stdout,
	}
}

// AddBar adds a progress bar
func (mpb *MultiProgressBar) AddBar(total int64, message string) *ProgressBar {
	mpb.mu.Lock()
	defer mpb.mu.Unlock()

	bar := NewProgressBar(total, message)
	bar.SetWriter(mpb.writer)
	mpb.bars = append(mpb.bars, bar)
	return bar
}

// Render renders all progress bars
func (mpb *MultiProgressBar) Render() {
	mpb.mu.Lock()
	defer mpb.mu.Unlock()

	// Move cursor up to overwrite previous output
	if len(mpb.bars) > 1 {
		fmt.Fprintf(mpb.writer, "\033[%dA", len(mpb.bars))
	}

	for i, bar := range mpb.bars {
		bar.render()
		if i < len(mpb.bars)-1 {
			fmt.Fprintln(mpb.writer)
		}
	}
}

// Specialized progress bars for different use cases

// ZKProofProgress creates a ZK proof generation progress bar
func ZKProofProgress(circuit string, phases []string) *MultiProgressBar {
	mpb := NewMultiProgressBar()

	for _, phase := range phases {
		bar := mpb.AddBar(100, fmt.Sprintf("🔐 %s - %s", circuit, phase))
		bar.SetColor(currentScheme.Accent)
		bar.SetChars("▓", "░")
	}

	return mpb
}

// ReserveMonitorProgress creates a reserve monitoring progress bar
func ReserveMonitorProgress(message string) *ProgressBar {
	bar := NewProgressBar(100, fmt.Sprintf("📊 %s", message))
	bar.SetColor(currentScheme.Success)
	bar.SetChars("█", "▒")
	bar.SetShowRate(true, "checks")
	return bar
}

// Utility functions
func formatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	} else if n < 1000000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	} else if n < 1000000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	} else {
		return fmt.Sprintf("%.1fB", float64(n)/1000000000)
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	} else {
		return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
	}
}

// Convenience function for simple progress tracking
func TrackProgress(total int64, message string, fn func(update func(int64))) {
	bar := NewProgressBar(total, message).SetShowETA(true)

	updateFn := func(current int64) {
		bar.Update(current)
	}

	fn(updateFn)
	bar.Finish()
}
