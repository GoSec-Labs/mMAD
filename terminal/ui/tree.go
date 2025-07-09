package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// TreeStyle defines tree appearance
type TreeStyle struct {
	Branch    string
	LastItem  string
	Pipe      string
	Space     string
	IconColor string
	TextColor string
}

// Predefined tree styles
var (
	DefaultTreeStyle = TreeStyle{
		Branch:    "├── ",
		LastItem:  "└── ",
		Pipe:      "│   ",
		Space:     "    ",
		IconColor: currentScheme.Primary,
		TextColor: "",
	}

	ASCIITreeStyle = TreeStyle{
		Branch:    "|-- ",
		LastItem:  "`-- ",
		Pipe:      "|   ",
		Space:     "    ",
		IconColor: currentScheme.Primary,
		TextColor: "",
	}

	MinimalTreeStyle = TreeStyle{
		Branch:    "- ",
		LastItem:  "- ",
		Pipe:      "  ",
		Space:     "  ",
		IconColor: currentScheme.Muted,
		TextColor: "",
	}

	FinancialTreeStyle = TreeStyle{
		Branch:    "┣━━ ",
		LastItem:  "┗━━ ",
		Pipe:      "┃   ",
		Space:     "    ",
		IconColor: currentScheme.Accent,
		TextColor: "",
	}
)

// TreeNode represents a node in the tree
type TreeNode struct {
	Label       string
	Icon        string
	Color       string
	Description string
	Data        interface{}
	Children    []*TreeNode
	Expanded    bool
	Metadata    map[string]interface{}
}

// Tree represents a tree structure
type Tree struct {
	root     *TreeNode
	style    TreeStyle
	writer   io.Writer
	title    string
	maxDepth int
}

// NewTree creates a new tree
func NewTree(rootLabel string) *Tree {
	return &Tree{
		root: &TreeNode{
			Label:    rootLabel,
			Children: make([]*TreeNode, 0),
			Expanded: true,
		},
		style:    DefaultTreeStyle,
		writer:   os.Stdout,
		maxDepth: -1, // No limit
	}
}

// SetStyle sets the tree style
func (t *Tree) SetStyle(style TreeStyle) *Tree {
	t.style = style
	return t
}

// SetWriter sets the output writer
func (t *Tree) SetWriter(writer io.Writer) *Tree {
	t.writer = writer
	return t
}

// SetTitle sets the tree title
func (t *Tree) SetTitle(title string) *Tree {
	t.title = title
	return t
}

// SetMaxDepth sets maximum display depth
func (t *Tree) SetMaxDepth(depth int) *Tree {
	t.maxDepth = depth
	return t
}

// GetRoot returns the root node
func (t *Tree) GetRoot() *TreeNode {
	return t.root
}

// AddChild adds a child to the root
func (t *Tree) AddChild(label string) *TreeNode {
	return t.root.AddChild(label)
}

// AddChild adds a child node
func (n *TreeNode) AddChild(label string) *TreeNode {
	child := &TreeNode{
		Label:    label,
		Children: make([]*TreeNode, 0),
		Expanded: true,
		Metadata: make(map[string]interface{}),
	}
	n.Children = append(n.Children, child)
	return child
}

// AddChildWithIcon adds a child with an icon
func (n *TreeNode) AddChildWithIcon(label, icon string) *TreeNode {
	child := n.AddChild(label)
	child.Icon = icon
	return child
}

// AddChildWithColor adds a child with custom color
func (n *TreeNode) AddChildWithColor(label, color string) *TreeNode {
	child := n.AddChild(label)
	child.Color = color
	return child
}

// AddChildWithDetails adds a child with full details
func (n *TreeNode) AddChildWithDetails(label, icon, description, color string) *TreeNode {
	child := n.AddChild(label)
	child.Icon = icon
	child.Description = description
	child.Color = color
	return child
}

// SetExpanded sets node expansion state
func (n *TreeNode) SetExpanded(expanded bool) *TreeNode {
	n.Expanded = expanded
	return n
}

// SetData sets node data
func (n *TreeNode) SetData(data interface{}) *TreeNode {
	n.Data = data
	return n
}

// SetMetadata sets node metadata
func (n *TreeNode) SetMetadata(key string, value interface{}) *TreeNode {
	if n.Metadata == nil {
		n.Metadata = make(map[string]interface{})
	}
	n.Metadata[key] = value
	return n
}

// Render renders the tree
func (t *Tree) Render() {
	if t.title != "" {
		fmt.Fprintf(t.writer, "\n%s\n", Primary(Bold+t.title))
	}

	t.renderNode(t.root, "", true, 0)
	fmt.Fprintln(t.writer)
}

// renderNode renders a single node and its children
func (t *Tree) renderNode(node *TreeNode, prefix string, isLast bool, depth int) {
	if t.maxDepth >= 0 && depth > t.maxDepth {
		return
	}

	// Don't render root node label if it's empty
	if depth > 0 || node.Label != "" {
		// Determine the connector
		var connector string
		if depth == 0 {
			connector = ""
		} else if isLast {
			connector = colorize(t.style.IconColor, t.style.LastItem)
		} else {
			connector = colorize(t.style.IconColor, t.style.Branch)
		}

		// Build the line
		var line strings.Builder
		line.WriteString(prefix)
		line.WriteString(connector)

		// Add icon
		if node.Icon != "" {
			icon := colorize(t.style.IconColor, node.Icon+" ")
			line.WriteString(icon)
		}

		// Add label with color
		label := node.Label
		if node.Color != "" {
			label = colorize(node.Color, label)
		} else if t.style.TextColor != "" {
			label = colorize(t.style.TextColor, label)
		}
		line.WriteString(label)

		// Add description
		if node.Description != "" {
			desc := colorize(t.style.IconColor, " - "+node.Description)
			line.WriteString(desc)
		}

		// Add metadata
		if len(node.Metadata) > 0 {
			var meta []string
			for k, v := range node.Metadata {
				meta = append(meta, fmt.Sprintf("%s: %v", k, v))
			}
			metaStr := colorize(Dim, " ("+strings.Join(meta, ", ")+")")
			line.WriteString(metaStr)
		}

		fmt.Fprintln(t.writer, line.String())
	}

	// Render children if expanded
	if node.Expanded {
		for i, child := range node.Children {
			isLastChild := i == len(node.Children)-1

			// Build prefix for children
			var childPrefix string
			if depth == 0 && node.Label == "" {
				childPrefix = prefix
			} else if isLast {
				childPrefix = prefix + colorize(t.style.IconColor, t.style.Space)
			} else {
				childPrefix = prefix + colorize(t.style.IconColor, t.style.Pipe)
			}

			t.renderNode(child, childPrefix, isLastChild, depth+1)
		}
	}
}

// FindNode finds a node by label (breadth-first search)
func (t *Tree) FindNode(label string) *TreeNode {
	return findNodeRecursive(t.root, label)
}

func findNodeRecursive(node *TreeNode, label string) *TreeNode {
	if node.Label == label {
		return node
	}

	for _, child := range node.Children {
		if found := findNodeRecursive(child, label); found != nil {
			return found
		}
	}

	return nil
}

// Specialized tree builders

// NewFileTree creates a file system tree
func NewFileTree(rootPath string) *Tree {
	tree := NewTree(rootPath)
	tree.SetStyle(DefaultTreeStyle)
	tree.SetTitle("📁 File Structure")
	return tree
}

// NewCircuitTree creates a ZK circuit dependency tree
func NewCircuitTree() *Tree {
	tree := NewTree("🔐 ZK Circuits")
	tree.SetStyle(FinancialTreeStyle)
	return tree
}

// NewReserveTree creates a reserve structure tree
func NewReserveTree() *Tree {
	tree := NewTree("💰 Reserve Structure")
	tree.SetStyle(DefaultTreeStyle)
	return tree
}

// NewComplianceTree creates a compliance hierarchy tree
func NewComplianceTree() *Tree {
	tree := NewTree("📋 Compliance Framework")
	tree.SetStyle(MinimalTreeStyle)
	return tree
}

// Helper functions for common tree patterns

// AddFileNode adds a file node with appropriate icon
func AddFileNode(parent *TreeNode, filename, fileType string) *TreeNode {
	var icon string
	var color string

	switch fileType {
	case "directory", "folder":
		icon = "📁"
		color = Primary("")
	case "go":
		icon = "🔵"
		color = Cyan
	case "json":
		icon = "📄"
		color = Yellow
	case "yaml", "yml":
		icon = "📋"
		color = Blue
	case "md":
		icon = "📝"
		color = Green
	case "txt":
		icon = "📃"
		color = Muted("")
	case "circuit":
		icon = "🔐"
		color = Magenta
	case "key":
		icon = "🔑"
		color = Red
	default:
		icon = "📄"
		color = ""
	}

	return parent.AddChildWithDetails(filename, icon, fileType, color)
}

// AddStatusNode adds a status node with appropriate styling
func AddStatusNode(parent *TreeNode, label, status string) *TreeNode {
	var icon string
	var color string

	switch strings.ToLower(status) {
	case "active", "running", "success", "completed":
		icon = "✅"
		color = Success("")
	case "error", "failed", "stopped":
		icon = "❌"
		color = Error("")
	case "warning", "pending":
		icon = "⚠️"
		color = Warning("")
	case "info", "loading":
		icon = "🔵"
		color = Info("")
	default:
		icon = "⚪"
		color = Muted("")
	}

	return parent.AddChildWithDetails(label, icon, status, color)
}

// Quick tree building functions

// QuickTree builds a simple tree from nested map
func QuickTree(title string, data map[string]interface{}) *Tree {
	tree := NewTree("")
	tree.SetTitle(title)

	buildFromMap(tree.GetRoot(), data)
	return tree
}

func buildFromMap(parent *TreeNode, data map[string]interface{}) {
	for key, value := range data {
		node := parent.AddChild(key)

		switch v := value.(type) {
		case map[string]interface{}:
			buildFromMap(node, v)
		case []interface{}:
			for i, item := range v {
				itemNode := node.AddChild(fmt.Sprintf("[%d]", i))
				if itemMap, ok := item.(map[string]interface{}); ok {
					buildFromMap(itemNode, itemMap)
				} else {
					itemNode.AddChild(fmt.Sprintf("%v", item))
				}
			}
		default:
			node.Description = fmt.Sprintf("%v", v)
		}
	}
}

// TreeFromStructure creates a tree from a simple structure
func TreeFromStructure(title string, structure []string) *Tree {
	tree := NewTree("")
	tree.SetTitle(title)

	for _, item := range structure {
		parts := strings.Split(item, "/")
		current := tree.GetRoot()

		for _, part := range parts {
			if part == "" {
				continue
			}

			// Find existing child or create new one
			var found *TreeNode
			for _, child := range current.Children {
				if child.Label == part {
					found = child
					break
				}
			}

			if found == nil {
				found = current.AddChild(part)
			}

			current = found
		}
	}

	return tree
}
