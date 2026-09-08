package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"time"
)

type Client struct {
	httpClient *http.Client
	apiKey     string
}

const (
	Dimensions = 768
	endpoint   = "https://generativelanguage.googleapis.com/v1beta/models/gemini-embedding-001:embedContent"
)

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		apiKey:     os.Getenv("GEMINI_API_KEY"),
	}
}

func (c *Client) Enabled() bool { return c.apiKey != "" }

type embedContentPart struct {
	Text string `json:"text"`
}

type embedContent struct {
	Parts []embedContentPart `json:"parts"`
}

type embedRequest struct {
	Model   string       `json:"model"`
	Content embedContent `json:"content"`

	OutputDimensionality int `json:"output_dimensionality"`
}

type embedResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
}

func (c *Client) Embed(ctx context.Context, text string) ([]float32, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("gemini embeddings: GEMINI_API_KEY not set")
	}

	reqData := embedRequest{
		Model:                "models/gemini-embedding-001",
		Content:              embedContent{Parts: []embedContentPart{{Text: text}}},
		OutputDimensionality: Dimensions,
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(reqData); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"?key="+c.apiKey, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini embeddings: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errBody)
		return nil, fmt.Errorf("gemini embeddings: API error (status %d): %v", resp.StatusCode, errBody)
	}

	var parsed embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	normalize(parsed.Embedding.Values)
	return parsed.Embedding.Values, nil
}

func normalize(v []float32) {
	var sumSquares float64
	for _, x := range v {
		sumSquares += float64(x) * float64(x)
	}
	norm := math.Sqrt(sumSquares)
	if norm == 0 {
		return
	}
	for i := range v {
		v[i] = float32(float64(v[i]) / norm)
	}
}
