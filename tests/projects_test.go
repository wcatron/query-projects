package commands

import (
	"testing"
)

func TestExtractTypeScriptCode_ValidTypeScriptBlock(t *testing.T) {
	response := "Here is some code:\n```typescript\nconst x = 10;\n```"
	expected := "const x = 10;"
	result := extractTypeScriptCode(response)
	if result != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, result)
	}
}

func TestExtractTypeScriptCode_ValidTSBlock(t *testing.T) {
	response := "Here is some code:\n```ts\nlet y = 20;\n```"
	expected := "let y = 20;"
	result := extractTypeScriptCode(response)
	if result != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, result)
	}
}

func TestExtractTypeScriptCode_NoTypeScriptBlock(t *testing.T) {
	response := "Here is some text without code blocks."
	expected := ""
	result := extractTypeScriptCode(response)
	if result != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, result)
	}
}

func TestExtractTypeScriptCode_EmptyTypeScriptBlock(t *testing.T) {
	response := "Here is an empty code block:\n```typescript\n\n```"
	expected := ""
	result := extractTypeScriptCode(response)
	if result != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, result)
	}
}
