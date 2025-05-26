package graphs

import (
	"fmt"
	"strings"
	"testing"
)

func TestGetDeleteMessages(t *testing.T) {
	existingMemories := "memory1, memory2"
	data := "delete memory1"
	userID := "test_user_123"

	systemPrompt, userPrompt := GetDeleteMessages(existingMemories, data, userID)

	// Verify systemPrompt
	if !strings.Contains(systemPrompt, userID) {
		t.Errorf("Expected systemPrompt to contain userID '%s', but it did not. Got: %s", userID, systemPrompt)
	}
	if strings.Contains(systemPrompt, "USER_ID") {
		t.Errorf("Expected systemPrompt to have 'USER_ID' replaced, but it was found. Got: %s", systemPrompt)
	}

	// Verify userPrompt
	expectedUserPrompt := fmt.Sprintf("Here are the existing memories: %s \n\n New Information: %s", existingMemories, data)
	if userPrompt != expectedUserPrompt {
		t.Errorf("Expected userPrompt to be '%s', got '%s'", expectedUserPrompt, userPrompt)
	}
}

func TestPromptConstants(t *testing.T) {
	t.Run("UpdateGraphPromptTemplate not empty", func(t *testing.T) {
		if UpdateGraphPromptTemplate == "" {
			t.Error("Expected UpdateGraphPromptTemplate to not be empty, but it was.")
		}
	})

	t.Run("ExtractRelationsPromptTemplate not empty", func(t *testing.T) {
		if ExtractRelationsPromptTemplate == "" {
			t.Error("Expected ExtractRelationsPromptTemplate to not be empty, but it was.")
		}
		if !strings.Contains(ExtractRelationsPromptTemplate, "CUSTOM_PROMPT") {
			t.Error("Expected ExtractRelationsPromptTemplate to contain CUSTOM_PROMPT placeholder.")
		}
	})

	t.Run("DeleteRelationsSystemPromptTemplate not empty", func(t *testing.T) {
		if DeleteRelationsSystemPromptTemplate == "" {
			t.Error("Expected DeleteRelationsSystemPromptTemplate to not be empty, but it was.")
		}
		if !strings.Contains(DeleteRelationsSystemPromptTemplate, "USER_ID") {
			t.Error("Expected DeleteRelationsSystemPromptTemplate to contain USER_ID placeholder.")
		}
	})
}
