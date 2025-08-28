package parser

import (
	"testing"
)

func TestParseTrigger(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantTrigger string
		wantParams  map[string]string
		wantErr     bool
	}{
		{
			name:        "simple trigger",
			input:       ":ty",
			wantTrigger: ":ty",
			wantParams:  map[string]string{},
			wantErr:     false,
		},
		{
			name:        "trigger with single param",
			input:       ":ty?lang=vi",
			wantTrigger: ":ty",
			wantParams:  map[string]string{"lang": "vi"},
			wantErr:     false,
		},
		{
			name:        "trigger with multiple params",
			input:       ":ty?lang=vi&tone=casual",
			wantTrigger: ":ty",
			wantParams:  map[string]string{"lang": "vi", "tone": "casual"},
			wantErr:     false,
		},
		{
			name:        "trigger with empty param value",
			input:       ":date?format=",
			wantTrigger: ":date",
			wantParams:  map[string]string{"format": ""},
			wantErr:     false,
		},
		{
			name:        "trigger with URL encoded params",
			input:       ":date?format=Mon%2C%2002%20Jan%202006",
			wantTrigger: ":date",
			wantParams:  map[string]string{"format": "Mon, 02 Jan 2006"},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTrigger(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTrigger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Trigger != tt.wantTrigger {
				t.Errorf("ParseTrigger() trigger = %v, want %v", got.Trigger, tt.wantTrigger)
			}
			if len(got.Params) != len(tt.wantParams) {
				t.Errorf("ParseTrigger() params length = %v, want %v", len(got.Params), len(tt.wantParams))
				return
			}
			for key, wantValue := range tt.wantParams {
				if gotValue, exists := got.Params[key]; !exists || gotValue != wantValue {
					t.Errorf("ParseTrigger() params[%s] = %v, want %v", key, gotValue, wantValue)
				}
			}
		})
	}
}

func TestMergeParams(t *testing.T) {
	tests := []struct {
		name            string
		queryParams     map[string]string
		snippetDefaults map[string]any
		globalDefaults  map[string]any
		want            map[string]any
	}{
		{
			name:            "query params override snippet defaults",
			queryParams:     map[string]string{"lang": "vi"},
			snippetDefaults: map[string]any{"lang": "en", "tone": "neutral"},
			globalDefaults:  map[string]any{"format": "2006-01-02"},
			want:            map[string]any{"lang": "vi", "tone": "neutral", "format": "2006-01-02"},
		},
		{
			name:            "snippet defaults override global defaults",
			queryParams:     map[string]string{},
			snippetDefaults: map[string]any{"format": "Mon, 02 Jan 2006"},
			globalDefaults:  map[string]any{"format": "2006-01-02", "timezone": "UTC"},
			want:            map[string]any{"format": "Mon, 02 Jan 2006", "timezone": "UTC"},
		},
		{
			name:            "boolean conversion",
			queryParams:     map[string]string{"upper": "true", "enabled": "false"},
			snippetDefaults: map[string]any{},
			globalDefaults:  map[string]any{},
			want:            map[string]any{"upper": true, "enabled": false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeParams(tt.queryParams, tt.snippetDefaults, tt.globalDefaults)
			for key, wantValue := range tt.want {
				if gotValue, exists := got[key]; !exists || gotValue != wantValue {
					t.Errorf("MergeParams()[%s] = %v, want %v", key, gotValue, wantValue)
				}
			}
		})
	}
}

func TestValidateTrigger(t *testing.T) {
	tests := []struct {
		name    string
		trigger string
		want    bool
	}{
		{"valid trigger", ":ty", true},
		{"valid trigger with prefix", "@hello", true},
		{"empty trigger", "", false},
		{"trigger with space", ":hello world", false},
		{"trigger with tab", ":hello\tworld", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateTrigger(tt.trigger); got != tt.want {
				t.Errorf("ValidateTrigger() = %v, want %v", got, tt.want)
			}
		})
	}
}
