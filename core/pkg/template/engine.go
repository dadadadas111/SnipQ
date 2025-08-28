package template

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
)

// Engine handles template rendering with built-in functions
type Engine struct {
	template *template.Template
}

// NewEngine creates a new template engine with built-in functions
func NewEngine() *Engine {
	tmpl := template.New("snipq").Funcs(template.FuncMap{
		"date":      dateFunc,
		"uuid":      uuidFunc,
		"counter":   counterFunc,
		"clipboard": clipboardFunc,
		"random":    randomFunc,
		"upper":     strings.ToUpper,
		"lower":     strings.ToLower,
		"title":     titleFunc,
		"trim":      strings.TrimSpace,
		"eq":        equal,
		"ne":        notEqual,
		"lt":        lessThan,
		"le":        lessEqual,
		"gt":        greaterThan,
		"ge":        greaterEqual,
	})

	return &Engine{
		template: tmpl,
	}
}

// Render renders a template with the given data
func (e *Engine) Render(templateText string, data map[string]any) (string, error) {
	tmpl, err := e.template.Clone()
	if err != nil {
		return "", err
	}

	tmpl, err = tmpl.Parse(templateText)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}

	return buf.String(), nil
}

// Built-in template functions

// dateFunc formats the current date/time
func dateFunc(format, timezone string) string {
	var loc *time.Location
	var err error

	switch timezone {
	case "Local", "":
		loc = time.Local
	case "UTC":
		loc = time.UTC
	default:
		loc, err = time.LoadLocation(timezone)
		if err != nil {
			loc = time.Local
		}
	}

	return time.Now().In(loc).Format(format)
}

// uuidFunc generates a UUID
func uuidFunc(withHyphens bool) string {
	id := uuid.New()
	if withHyphens {
		return id.String()
	}
	return strings.ReplaceAll(id.String(), "-", "")
}

// counterFunc increments and returns a counter value
// This is a placeholder - actual implementation needs counter state management
func counterFunc(name string, pad int) string {
	// TODO: Implement actual counter management
	// For now, return a placeholder
	return fmt.Sprintf("%0*d", pad, 1)
}

// clipboardFunc placeholder for clipboard content
func clipboardFunc() string {
	return ""
}

// titleFunc converts string to title case (replacement for deprecated strings.Title)
func titleFunc(s string) string {
	if s == "" {
		return s
	}

	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}

// randomFunc generates random values
func randomFunc(args ...any) string {
	if len(args) == 0 {
		// Generate random number 0-100
		n, _ := rand.Int(rand.Reader, big.NewInt(101))
		return n.String()
	}

	switch v := args[0].(type) {
	case string:
		if v == "word" {
			// Random word from a small set
			words := []string{"alpha", "beta", "gamma", "delta", "epsilon", "zeta"}
			n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(words))))
			return words[n.Int64()]
		}
	case int:
		// Random number 0 to v
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(v+1)))
		return n.String()
	}

	return "random"
}

// Comparison functions for templates

func equal(a, b any) bool {
	return a == b
}

func notEqual(a, b any) bool {
	return a != b
}

func lessThan(a, b any) bool {
	return compareValues(a, b) < 0
}

func lessEqual(a, b any) bool {
	return compareValues(a, b) <= 0
}

func greaterThan(a, b any) bool {
	return compareValues(a, b) > 0
}

func greaterEqual(a, b any) bool {
	return compareValues(a, b) >= 0
}

func compareValues(a, b any) int {
	// Simple string comparison for now
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return strings.Compare(aStr, bStr)
}
