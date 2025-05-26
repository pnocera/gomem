package memory

import (
	"strings"
	"testing"
	"time"
)

func TestMessage_Validate(t *testing.T) {
	tests := []struct {
		name    string
		msg     Message
		wantErr bool
		errText string // Substring to check in error message
	}{
		{"Valid User Message", Message{Role: "user", Content: "Hello"}, false, ""},
		{"Valid Assistant Message", Message{Role: "assistant", Content: "Hi there"}, false, ""},
		{"Valid System Message", Message{Role: "system", Content: "System init"}, false, ""},
		{"Invalid Role", Message{Role: "invalid_role", Content: "Test"}, true, "Key: 'Message.Role' Error:Field validation for 'Role' failed on the 'oneof' tag"},
		{"Missing Content", Message{Role: "user", Content: ""}, true, "Key: 'Message.Content' Error:Field validation for 'Content' failed on the 'required' tag"},
		{"Missing Role", Message{Role: "", Content: "Test"}, true, "Key: 'Message.Role' Error:Field validation for 'Role' failed on the 'required' tag"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Message.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errText != "" {
				if !strings.Contains(err.Error(), tt.errText) {
					t.Errorf("Message.Validate() error = %v, wantErrText %s", err, tt.errText)
				}
			}
		})
	}
}

func TestAddMemoryRequest_Validate(t *testing.T) {
	validMsg := Message{Role: "user", Content: "Valid message"}
	invalidMsg_NoRole := Message{Content: "Invalid - no role"}
	invalidMsg_NoContent := Message{Role: "user"}

	tests := []struct {
		name    string
		req     AddMemoryRequest
		wantErr bool
		errText string
	}{
		{"Valid Request", AddMemoryRequest{Messages: []Message{validMsg}}, false, ""},
		{"Missing Messages", AddMemoryRequest{}, true, "Key: 'AddMemoryRequest.Messages' Error:Field validation for 'Messages' failed on the 'required' tag"},
		{"Empty Messages Slice", AddMemoryRequest{Messages: []Message{}}, true, "Key: 'AddMemoryRequest.Messages' Error:Field validation for 'Messages' failed on the 'min' tag"},
		{"Messages with Invalid Message (No Role)", AddMemoryRequest{Messages: []Message{validMsg, invalidMsg_NoRole}}, true, "Key: 'AddMemoryRequest.Messages[1].Role' Error:Field validation for 'Role' failed on the 'required' tag"},
		{"Messages with Invalid Message (No Content)", AddMemoryRequest{Messages: []Message{validMsg, invalidMsg_NoContent}}, true, "Key: 'AddMemoryRequest.Messages[1].Content' Error:Field validation for 'Content' failed on the 'required' tag"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("AddMemoryRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errText != "" {
				if !strings.Contains(err.Error(), tt.errText) {
					t.Errorf("AddMemoryRequest.Validate() error = %v, wantErrText %s", err, tt.errText)
				}
			}
		})
	}
}

func TestSearchMemoryRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     SearchMemoryRequest
		wantErr bool
		errText string
	}{
		{"Valid Request", SearchMemoryRequest{Query: "test query", Limit: 10}, false, ""},
		{"Valid Request No Limit", SearchMemoryRequest{Query: "test query"}, false, ""}, // Limit is optional
		{"Missing Query", SearchMemoryRequest{Limit: 10}, true, "Key: 'SearchMemoryRequest.Query' Error:Field validation for 'Query' failed on the 'required' tag"},
		{"Invalid Limit (Zero)", SearchMemoryRequest{Query: "test", Limit: 0}, true, "Key: 'SearchMemoryRequest.Limit' Error:Field validation for 'Limit' failed on the 'gt' tag"},
		{"Invalid Limit (Negative)", SearchMemoryRequest{Query: "test", Limit: -5}, true, "Key: 'SearchMemoryRequest.Limit' Error:Field validation for 'Limit' failed on the 'gt' tag"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchMemoryRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errText != "" {
				if !strings.Contains(err.Error(), tt.errText) {
					t.Errorf("SearchMemoryRequest.Validate() error = %v, wantErrText %s", err, tt.errText)
				}
			}
		})
	}
}

func TestMemoryEvent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		event   MemoryEvent
		wantErr bool
		errText string
	}{
		{"Valid Event", MemoryEvent{EventType: "TEST_EVENT", Timestamp: time.Now()}, false, ""},
		{"Missing EventType", MemoryEvent{Timestamp: time.Now()}, true, "Key: 'MemoryEvent.EventType' Error:Field validation for 'EventType' failed on the 'required' tag"},
		// Timestamp is not validated by struct tag, but usually set by system.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("MemoryEvent.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errText != "" {
				if !strings.Contains(err.Error(), tt.errText) {
					t.Errorf("MemoryEvent.Validate() error = %v, wantErrText %s", err, tt.errText)
				}
			}
		})
	}
}
