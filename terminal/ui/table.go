package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

// TableStyle defines table appearance
type TableStyle struct {
	Border      bool
	HeaderColor string
	RowColor    string
	BorderColor string
	Padding     int
	Separator   string
}

// Predefined table styles
var (
	DefaultTableStyle = TableStyle{
		Border:      true,
		HeaderColor: currentScheme.Primary,
		RowColor:    "",
		BorderColor: currentScheme.Muted,
		Padding:     1,
		Separator:   "│",
	}

	MinimalTableStyle = TableStyle{
		Border:      false,
		HeaderColor: currentScheme.Primary,
		RowColor:    "",
		BorderColor: "",
		Padding:     2,
		Separator:   " ",
	}

	FinancialTableStyle = TableStyle{
		Border:      true,
		HeaderColor: currentScheme.Primary + Bold,
		RowColor:    "",
		BorderColor: currentScheme.Secondary,
		Padding:     1,
		Separator:   "┃",
	}
)

// Column represents a table column
type Column struct {
	Header    string
	Field     string
	Width     int
	Align     Alignment
	Color     string
	Formatter func(interface{}) string
}

// Alignment for table columns
type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

// Table represents a data table
type Table struct {
	columns []Column
	rows    []map[string]interface{}
	style   TableStyle
	writer  io.Writer
	title   string
}

// NewTable creates a new table
func NewTable() *Table {
	return &Table{
		columns: make([]Column, 0),
		rows:    make([]map[string]interface{}, 0),
		style:   DefaultTableStyle,
		writer:  os.Stdout,
	}
}

// SetStyle sets the table style
func (t *Table) SetStyle(style TableStyle) *Table {
	t.style = style
	return t
}

// SetWriter sets the output writer
func (t *Table) SetWriter(writer io.Writer) *Table {
	t.writer = writer
	return t
}

// SetTitle sets the table title
func (t *Table) SetTitle(title string) *Table {
	t.title = title
	return t
}

// AddColumn adds a column to the table
func (t *Table) AddColumn(header, field string) *Table {
	t.columns = append(t.columns, Column{
		Header: header,
		Field:  field,
		Width:  0, // Auto-calculate
		Align:  AlignLeft,
		Color:  "",
	})
	return t
}

// AddColumnWithOptions adds a column with custom options
func (t *Table) AddColumnWithOptions(header, field string, width int, align Alignment, color string) *Table {
	t.columns = append(t.columns, Column{
		Header: header,
		Field:  field,
		Width:  width,
		Align:  align,
		Color:  color,
	})
	return t
}

// AddFormattedColumn adds a column with custom formatter
func (t *Table) AddFormattedColumn(header, field string, formatter func(interface{}) string) *Table {
	t.columns = append(t.columns, Column{
		Header:    header,
		Field:     field,
		Width:     0,
		Align:     AlignLeft,
		Color:     "",
		Formatter: formatter,
	})
	return t
}

// AddRow adds a data row
func (t *Table) AddRow(data map[string]interface{}) *Table {
	t.rows = append(t.rows, data)
	return t
}

// AddRows adds multiple data rows
func (t *Table) AddRows(rows []map[string]interface{}) *Table {
	t.rows = append(t.rows, rows...)
	return t
}

// calculateWidths calculates column widths
func (t *Table) calculateWidths() {
	for i := range t.columns {
		col := &t.columns[i]

		if col.Width > 0 {
			continue // Use specified width
		}

		// Start with header width
		col.Width = utf8.RuneCountInString(col.Header)

		// Check all row values
		for _, row := range t.rows {
			if value, exists := row[col.Field]; exists {
				var cellText string
				if col.Formatter != nil {
					cellText = col.Formatter(value)
				} else {
					cellText = fmt.Sprintf("%v", value)
				}

				width := utf8.RuneCountInString(cellText)
				if width > col.Width {
					col.Width = width
				}
			}
		}

		// Add padding
		col.Width += t.style.Padding * 2
	}
}

// formatCell formats a cell value
func (t *Table) formatCell(value interface{}, col Column) string {
	var text string

	if col.Formatter != nil {
		text = col.Formatter(value)
	} else {
		text = fmt.Sprintf("%v", value)
	}

	// Apply alignment
	padding := col.Width - utf8.RuneCountInString(text)
	if padding < 0 {
		padding = 0
	}

	switch col.Align {
	case AlignLeft:
		text = text + strings.Repeat(" ", padding)
	case AlignRight:
		text = strings.Repeat(" ", padding) + text
	case AlignCenter:
		leftPad := padding / 2
		rightPad := padding - leftPad
		text = strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
	}

	// Apply color
	if col.Color != "" {
		text = colorize(col.Color, text)
	}

	return text
}

// Render renders the table
func (t *Table) Render() {
	if len(t.columns) == 0 {
		return
	}

	t.calculateWidths()

	// Render title
	if t.title != "" {
		fmt.Fprintf(t.writer, "\n%s\n", Primary(Bold+t.title))
	}

	// Render top border
	if t.style.Border {
		t.renderBorder("top")
	}

	// Render header
	t.renderHeader()

	// Render header separator
	if t.style.Border {
		t.renderBorder("middle")
	}

	// Render rows
	for _, row := range t.rows {
		t.renderRow(row)
	}

	// Render bottom border
	if t.style.Border {
		t.renderBorder("bottom")
	}

	fmt.Fprintln(t.writer)
}

// renderBorder renders table borders
func (t *Table) renderBorder(position string) {
	var left, middle, right, horizontal string

	switch position {
	case "top":
		left, middle, right, horizontal = "┌", "┬", "┐", "─"
	case "middle":
		left, middle, right, horizontal = "├", "┼", "┤", "─"
	case "bottom":
		left, middle, right, horizontal = "└", "┴", "┘", "─"
	}

	var border strings.Builder
	border.WriteString(colorize(t.style.BorderColor, left))

	for i, col := range t.columns {
		border.WriteString(colorize(t.style.BorderColor, strings.Repeat(horizontal, col.Width)))
		if i < len(t.columns)-1 {
			border.WriteString(colorize(t.style.BorderColor, middle))
		}
	}

	border.WriteString(colorize(t.style.BorderColor, right))
	fmt.Fprintln(t.writer, border.String())
}

// renderHeader renders the table header
func (t *Table) renderHeader() {
	var row strings.Builder

	if t.style.Border {
		row.WriteString(colorize(t.style.BorderColor, t.style.Separator))
	}

	for i, col := range t.columns {
		headerText := t.formatCell(col.Header, Column{
			Width: col.Width,
			Align: col.Align,
			Color: t.style.HeaderColor,
		})

		row.WriteString(headerText)

		if i < len(t.columns)-1 || t.style.Border {
			row.WriteString(colorize(t.style.BorderColor, t.style.Separator))
		}
	}

	fmt.Fprintln(t.writer, row.String())
}

// renderRow renders a data row
func (t *Table) renderRow(data map[string]interface{}) {
	var row strings.Builder

	if t.style.Border {
		row.WriteString(colorize(t.style.BorderColor, t.style.Separator))
	}

	for i, col := range t.columns {
		value := data[col.Field]
		cellText := t.formatCell(value, col)

		if t.style.RowColor != "" {
			cellText = colorize(t.style.RowColor, cellText)
		}

		row.WriteString(cellText)

		if i < len(t.columns)-1 || t.style.Border {
			row.WriteString(colorize(t.style.BorderColor, t.style.Separator))
		}
	}

	fmt.Fprintln(t.writer, row.String())
}

// Financial-specific table helpers

// NewReserveTable creates a table for reserve data
func NewReserveTable() *Table {
	return NewTable().
		SetStyle(FinancialTableStyle).
		SetTitle("💰 Reserve Status").
		AddColumn("Asset", "asset").
		AddFormattedColumn("Balance", "balance", FormatCurrency).
		AddFormattedColumn("Value (USD)", "value_usd", FormatCurrency).
		AddFormattedColumn("Ratio", "ratio", FormatPercentage)
}

// NewProofTable creates a table for ZK proof data
func NewProofTable() *Table {
	return NewTable().
		SetStyle(DefaultTableStyle).
		SetTitle("🔐 ZK Proofs").
		AddColumn("Circuit", "circuit").
		AddColumn("Status", "status").
		AddFormattedColumn("Generated", "timestamp", FormatTimestamp).
		AddColumn("Verified", "verified")
}

// NewTransactionTable creates a table for transaction data
func NewTransactionTable() *Table {
	return NewTable().
		SetStyle(MinimalTableStyle).
		SetTitle("📊 Recent Transactions").
		AddColumn("Hash", "hash").
		AddFormattedColumn("Amount", "amount", FormatCurrency).
		AddColumn("Type", "type").
		AddFormattedColumn("Time", "timestamp", FormatTimestamp)
}

// Formatter functions
func FormatCurrency(value interface{}) string {
	if v, ok := value.(float64); ok {
		if v >= 1000000 {
			return fmt.Sprintf("$%.2fM", v/1000000)
		} else if v >= 1000 {
			return fmt.Sprintf("$%.2fK", v/1000)
		}
		return fmt.Sprintf("$%.2f", v)
	}
	return fmt.Sprintf("%v", value)
}

func FormatPercentage(value interface{}) string {
	if v, ok := value.(float64); ok {
		color := Success("")
		if v < 100 {
			color = Warning("")
		}
		if v < 90 {
			color = Error("")
		}
		return colorize(color, fmt.Sprintf("%.1f%%", v))
	}
	return fmt.Sprintf("%v%%", value)
}

func FormatTimestamp(value interface{}) string {
	if v, ok := value.(string); ok {
		// Truncate long timestamps
		if len(v) > 16 {
			return v[:16]
		}
		return v
	}
	return fmt.Sprintf("%v", value)
}

// Quick table creation functions
func QuickTable(headers []string, rows [][]string) {
	table := NewTable()

	for i, header := range headers {
		table.AddColumn(header, fmt.Sprintf("col_%d", i))
	}

	for _, row := range rows {
		data := make(map[string]interface{})
		for i, cell := range row {
			data[fmt.Sprintf("col_%d", i)] = cell
		}
		table.AddRow(data)
	}

	table.Render()
}
