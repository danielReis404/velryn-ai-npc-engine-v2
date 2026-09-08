package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestActionVerbRoundTripsThroughJSON(t *testing.T) {
	action := AgentAction{
		InternalThought: "I should retreat.",
		Action:          ActionFlee,
	}

	data, err := json.Marshal(action)
	if err != nil {
		t.Fatalf("failed to marshal AgentAction: %v", err)
	}
	if want := `"action":"FLEE"`; !strings.Contains(string(data), want) {
		t.Errorf("expected marshaled JSON to contain %s, got %s", want, data)
	}

	var decoded AgentAction
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal AgentAction: %v", err)
	}
	if decoded.Action != ActionFlee {
		t.Errorf("expected decoded action to equal ActionFlee, got %q", decoded.Action)
	}
}

func TestActionVerbAcceptsRawStringFromAgentService(t *testing.T) {

	raw := []byte(`{"internal_thought":"heading to the market","action":"MOVE"}`)

	var decoded AgentAction
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if decoded.Action != ActionMove {
		t.Errorf("expected ActionMove, got %q", decoded.Action)
	}
}
