package parser

import (
	"net/url"
	"strings"
)

// ParsedTrigger represents a parsed trigger with query parameters
type ParsedTrigger struct {
	Trigger string            `json:"trigger"`
	Params  map[string]string `json:"params"`
}

// ParseTrigger parses a raw trigger string into trigger and query parameters
// Example: ":ty?lang=vi&tone=casual" -> trigger=":ty", params={"lang":"vi", "tone":"casual"}
func ParseTrigger(rawTrigger string) (*ParsedTrigger, error) {
	// Split at the first ? to separate trigger from query params
	parts := strings.SplitN(rawTrigger, "?", 2)

	trigger := strings.TrimSpace(parts[0])
	params := make(map[string]string)

	// Parse query parameters if present
	if len(parts) > 1 {
		queryParams, err := url.ParseQuery(parts[1])
		if err != nil {
			return nil, err
		}

		// Convert url.Values to map[string]string (take first value for each key)
		for key, values := range queryParams {
			if len(values) > 0 {
				params[key] = values[0]
			}
		}
	}

	return &ParsedTrigger{
		Trigger: trigger,
		Params:  params,
	}, nil
}

// MergeParams merges query params with snippet defaults and global defaults
// Priority: query params > snippet defaults > global defaults
func MergeParams(queryParams map[string]string, snippetDefaults map[string]any, globalDefaults map[string]any) map[string]any {
	result := make(map[string]any)

	// Start with global defaults
	for key, value := range globalDefaults {
		result[key] = value
	}

	// Override with snippet defaults
	for key, value := range snippetDefaults {
		result[key] = value
	}

	// Override with query params (convert strings to appropriate types)
	for key, value := range queryParams {
		result[key] = convertStringValue(value)
	}

	return result
}

// convertStringValue converts string values to appropriate types
func convertStringValue(value string) any {
	// Handle boolean values
	switch strings.ToLower(value) {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	}

	// For now, keep as string - we can add more type conversion later
	return value
}

// NormalizeTrigger normalizes a trigger by removing extra whitespace
func NormalizeTrigger(trigger string) string {
	return strings.TrimSpace(trigger)
}

// ValidateTrigger validates that a trigger is properly formatted
func ValidateTrigger(trigger string) bool {
	if trigger == "" {
		return false
	}

	// Trigger should not contain whitespace characters
	if strings.ContainsAny(trigger, " \t\n\r") {
		return false
	}

	return true
}
