package store

import (
	"testing"
)

func TestParseMetadataValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		// Plain strings
		{"plain string", "hello", `"hello"`, false},
		{"string with spaces", "hello world", `"hello world"`, false},
		{"string with quotes", `say "hi"`, `"say \"hi\""`, false},
		{"empty string", "", `""`, false},
		{"string with special chars", "test@example.com", `"test@example.com"`, false},

		// Numbers (treated as strings)
		{"integer", "42", `"42"`, false},
		{"float", "3.14", `"3.14"`, false},
		{"negative", "-10", `"-10"`, false},
		{"scientific notation", "1e5", `"1e5"`, false},

		// Booleans (treated as strings)
		{"true", "true", `"true"`, false},
		{"false", "false", `"false"`, false},

		// Null (treated as string)
		{"null", "null", `"null"`, false},

		// Objects (preserved)
		{"empty object", "{}", `{}`, false},
		{"simple object", `{"a":"b"}`, `{"a":"b"}`, false},
		{"object with spaces", `{"key": "value"}`, `{"key": "value"}`, false},
		{"nested object", `{"x":{"y":"z"}}`, `{"x":{"y":"z"}}`, false},
		{"object with number", `{"count":42}`, `{"count":42}`, false},

		// Arrays (preserved)
		{"empty array", "[]", `[]`, false},
		{"number array", "[1,2,3]", `[1,2,3]`, false},
		{"string array", `["a","b"]`, `["a","b"]`, false},
		{"mixed array", `[1,"two",true]`, `[1,"two",true]`, false},
		{"nested array", `[[1,2],[3,4]]`, `[[1,2],[3,4]]`, false},

		// Whitespace handling
		{"leading space", " hello", `"hello"`, false},
		{"trailing space", "hello ", `"hello"`, false},
		{"spaces around object", " {} ", `{}`, false},
		{"spaces around array", " [] ", `[]`, false},
		{"multiple spaces", "  test  ", `"test"`, false},

		// Edge cases
		{"quoted string input", `"hello"`, `"\"hello\""`, false},
		{"incomplete json object", `{"incomplete`, `"{\"incomplete"`, false},
		{"incomplete json array", `[incomplete`, `"[incomplete"`, false},
		{"just braces", "}", `"}"`, false},
		{"just brackets", "]", `"]"`, false},
		{"unicode", "こんにちは", `"こんにちは"`, false},
		{"newline", "hello\nworld", `"hello\nworld"`, false},
		{"tab", "hello\tworld", `"hello\tworld"`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMetadataValue(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMetadataValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("parseMetadataValue() = %v, want %v", got, tt.expected)
			}
		})
	}
}
