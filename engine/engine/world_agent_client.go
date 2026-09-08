package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

func DecideWorldEventViaAgentService(ctx context.Context, snapshot WorldSnapshot) (*WorldEvent, error) {
	body, err := json.Marshal(snapshot)
	if err != nil {
		return nil, err
	}

	url := os.Getenv("AGENT_SERVICE_URL") + "/world-decide"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("agent_service (world-decide) failed: %w", err)
	}
	defer resp.Body.Close()

	var event WorldEvent
	if err := json.NewDecoder(resp.Body).Decode(&event); err != nil {
		return nil, fmt.Errorf("invalid json from world-decide: %w", err)
	}
	return &event, nil
}
