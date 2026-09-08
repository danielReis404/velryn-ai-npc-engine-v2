package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"ai-agent-engine/models"
)

const (
	agentRequestTimeout      = 30 * time.Second
	agentBatchRequestTimeout = 60 * time.Second
)

var agentHTTPClient = &http.Client{
	Timeout: agentBatchRequestTimeout,
}

func agentServiceURL(path string) string {
	return strings.TrimRight(os.Getenv("AGENT_SERVICE_URL"), "/") + path
}

func DecideViaAgentService(ctx context.Context, perception NPCPerception) (models.AgentAction, error) {
	body, err := json.Marshal(perception)
	if err != nil {
		return models.AgentAction{}, err
	}

	requestCtx, cancel := context.WithTimeout(ctx, agentRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		requestCtx,
		http.MethodPost,
		agentServiceURL("/decide"),
		bytes.NewReader(body),
	)
	if err != nil {
		return models.AgentAction{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := agentHTTPClient.Do(req)
	if err != nil {
		return models.AgentAction{}, fmt.Errorf("agent_service failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return models.AgentAction{}, fmt.Errorf(
			"agent_service retornou status %d: %s",
			resp.StatusCode,
			strings.TrimSpace(string(b)),
		)
	}

	var action models.AgentAction
	if err := json.NewDecoder(resp.Body).Decode(&action); err != nil {
		return models.AgentAction{}, fmt.Errorf("invalid json from agent_service: %w", err)
	}

	return action, nil
}

func DecideGroupViaAgentService(ctx context.Context, perceptions []NPCPerception) (map[string]models.AgentAction, error) {
	body, err := json.Marshal(map[string]interface{}{"members": perceptions})
	if err != nil {
		return nil, err
	}

	requestCtx, cancel := context.WithTimeout(ctx, agentBatchRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		requestCtx,
		http.MethodPost,
		agentServiceURL("/decide-batch"),
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := agentHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("agent_service: batch request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf(
			"agent_service returned status %d: %s",
			resp.StatusCode,
			strings.TrimSpace(string(b)),
		)
	}

	var result struct {
		Actions map[string]models.AgentAction `json:"actions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("agent_service: invalid batch response: %w", err)
	}

	return result.Actions, nil
}
