package template

import (
	"strings"
	"testing"
)

func TestEngine_Render(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name         string
		template     string
		data         map[string]any
		want         string
		wantErr      bool
		containsText string // for dynamic content like dates
	}{
		{
			name:     "simple text",
			template: "Hello, World!",
			data:     map[string]any{},
			want:     "Hello, World!",
			wantErr:  false,
		},
		{
			name:     "variable substitution",
			template: "Hello, {{ .name }}!",
			data:     map[string]any{"name": "Alice"},
			want:     "Hello, Alice!",
			wantErr:  false,
		},
		{
			name:         "date function",
			template:     "Today is {{ date \"2006-01-02\" \"Local\" }}",
			data:         map[string]any{},
			containsText: "Today is",
			wantErr:      false,
		},
		{
			name:     "uuid function",
			template: "ID: {{ uuid false }}",
			data:     map[string]any{},
			want:     "", // We'll check length instead
			wantErr:  false,
		},
		{
			name:     "upper function",
			template: "{{ upper \"hello\" }}",
			data:     map[string]any{},
			want:     "HELLO",
			wantErr:  false,
		},
		{
			name:     "conditional",
			template: "{{ if eq .lang \"vi\" }}Xin chào{{ else }}Hello{{ end }}",
			data:     map[string]any{"lang": "vi"},
			want:     "Xin chào",
			wantErr:  false,
		},
		{
			name:     "conditional false",
			template: "{{ if eq .lang \"vi\" }}Xin chào{{ else }}Hello{{ end }}",
			data:     map[string]any{"lang": "en"},
			want:     "Hello",
			wantErr:  false,
		},
		{
			name:     "invalid template",
			template: "{{ .invalid syntax",
			data:     map[string]any{},
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := engine.Render(tt.template, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Engine.Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if tt.containsText != "" {
				if !strings.Contains(got, tt.containsText) {
					t.Errorf("Engine.Render() = %v, should contain %v", got, tt.containsText)
				}
				return
			}

			if tt.name == "uuid function" {
				// UUID should be 32 characters (without hyphens) plus "ID: " prefix
				if len(got) != 36 || !strings.HasPrefix(got, "ID: ") {
					t.Errorf("Engine.Render() UUID result = %v, should be 'ID: ' + 32 chars", got)
				}
				return
			}

			if got != tt.want {
				t.Errorf("Engine.Render() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		name     string
		function string
		args     []any
		wantType string
	}{
		{
			name:     "date function returns string",
			function: "date",
			args:     []any{"2006-01-02", "Local"},
			wantType: "string",
		},
		{
			name:     "uuid function returns string",
			function: "uuid",
			args:     []any{false},
			wantType: "string",
		},
		{
			name:     "random function returns string",
			function: "random",
			args:     []any{},
			wantType: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()

			var template string
			switch tt.function {
			case "date":
				template = "{{ date \"2006-01-02\" \"Local\" }}"
			case "uuid":
				template = "{{ uuid false }}"
			case "random":
				template = "{{ random }}"
			}

			result, err := engine.Render(template, map[string]any{})
			if err != nil {
				t.Errorf("Function %s failed: %v", tt.function, err)
				return
			}

			if tt.wantType == "string" && result == "" {
				t.Errorf("Function %s returned empty string", tt.function)
			}
		})
	}
}
