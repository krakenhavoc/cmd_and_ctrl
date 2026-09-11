package model

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// anthropic.go is the one place in the bot that talks to a network.
//
// # Why raw net/http and not anthropic-sdk-go
//
// There is an official Go SDK and it is the right default for a new
// project. It is the wrong call for THIS one, and the reason is
// arithmetic rather than taste. `server/go.mod` has three direct
// dependencies and pins `go 1.22` with a comment saying why ("older
// collaborator installs stay supported"). anthropic-sdk-go requires
// `go 1.24` and brings the AWS SDK, the MCP SDK, jsonschema and gjson
// with it. Taking it would bump the language floor for the entire
// repository and multiply its dependency count, to make one
// single-turn POST with no tools, no streaming and no conversation
// state.
//
// So this is ~120 lines of encoding/json against a documented wire
// format, behind the Client interface, which is exactly where the
// swap happens if that trade ever changes: implement Client with the
// SDK and delete this file. Nothing else in the package moves.
//
// # What is deliberately not here
//
// No streaming (the reply is one small JSON object), no tools (the
// model selects an index, it does not act), no retries (the runner's
// MaxThink is a hard deadline — a retry spends the budget that
// Layer B's answer is already waiting inside), and no conversation
// history (each decision is independent; ADR 0033 §7's "stateless
// between games" applies within one too).

const (
	defaultEndpoint   = "https://api.anthropic.com/v1/messages"
	anthropicVersion  = "2023-06-01"
	defaultHTTPBudget = 10 * time.Second
)

// AnthropicClient calls the Messages API.
type AnthropicClient struct {
	// APIKey is the x-api-key credential. Required.
	APIKey string
	// Endpoint overrides the Messages API URL. Empty uses the
	// production endpoint.
	Endpoint string
	// HTTP is the client used for the call. Nil builds one with a
	// conservative timeout; the per-call deadline that actually
	// matters comes from ctx.
	HTTP *http.Client
}

// NewAnthropicClient reads the key from CMDCTRL_ANTHROPIC_API_KEY,
// falling back to ANTHROPIC_API_KEY, and returns nil when neither is
// set. A nil Client is not an error anywhere in this package — it is
// how a deployment without a key runs the bot on Layer A + B, which
// is a supported configuration and the one the outage drill proves.
func NewAnthropicClient() *AnthropicClient {
	key := os.Getenv("CMDCTRL_ANTHROPIC_API_KEY")
	if key == "" {
		key = os.Getenv("ANTHROPIC_API_KEY")
	}
	if key == "" {
		return nil
	}
	c := &AnthropicClient{APIKey: key}
	if ep := os.Getenv("CMDCTRL_ANTHROPIC_ENDPOINT"); ep != "" {
		c.Endpoint = ep
	}
	return c
}

// Name identifies the transport.
func (c *AnthropicClient) Name() string { return "anthropic" }

// --- wire types ----------------------------------------------------

type wireCacheControl struct {
	Type string `json:"type"`
}

type wireSystemBlock struct {
	Type         string            `json:"type"`
	Text         string            `json:"text"`
	CacheControl *wireCacheControl `json:"cache_control,omitempty"`
}

type wireMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type wireThinking struct {
	Type string `json:"type"`
}

type wireOutputConfig struct {
	Effort string `json:"effort,omitempty"`
}

type wireRequest struct {
	Model        string            `json:"model"`
	MaxTokens    int               `json:"max_tokens"`
	System       []wireSystemBlock `json:"system,omitempty"`
	Messages     []wireMessage     `json:"messages"`
	Thinking     *wireThinking     `json:"thinking,omitempty"`
	OutputConfig *wireOutputConfig `json:"output_config,omitempty"`
}

type wireContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type wireUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

type wireResponse struct {
	Model      string        `json:"model"`
	StopReason string        `json:"stop_reason"`
	Content    []wireContent `json:"content"`
	Usage      wireUsage     `json:"usage"`
	Error      *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// APIError is a non-2xx reply, kept typed so a log can tell a 401
// (misconfigured key — a deploy problem) from a 529 (overloaded — a
// weather problem).
type APIError struct {
	Status  int
	Kind    string
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("anthropic: %d %s: %s", e.Status, e.Kind, e.Message)
}

// Complete makes one Messages API call.
func (c *AnthropicClient) Complete(ctx context.Context, req Request) (Response, error) {
	if c == nil || c.APIKey == "" {
		return Response{}, ErrNoClient
	}
	body := wireRequest{
		Model:     req.Model,
		MaxTokens: req.MaxTokens,
		Messages:  []wireMessage{{Role: "user", Content: req.User}},
	}
	if body.MaxTokens <= 0 {
		body.MaxTokens = 256
	}
	for _, b := range req.System {
		blk := wireSystemBlock{Type: "text", Text: b.Text}
		if b.Cache {
			blk.CacheControl = &wireCacheControl{Type: "ephemeral"}
		}
		body.System = append(body.System, blk)
	}
	if req.Thinking != "" {
		body.Thinking = &wireThinking{Type: req.Thinking}
	}
	if req.Effort != "" {
		body.OutputConfig = &wireOutputConfig{Effort: req.Effort}
	}

	raw, err := json.Marshal(body)
	if err != nil {
		return Response{}, fmt.Errorf("anthropic: marshal request: %w", err)
	}
	endpoint := c.Endpoint
	if endpoint == "" {
		endpoint = defaultEndpoint
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return Response{}, fmt.Errorf("anthropic: build request: %w", err)
	}
	httpReq.Header.Set("content-type", "application/json")
	httpReq.Header.Set("x-api-key", c.APIKey)
	httpReq.Header.Set("anthropic-version", anthropicVersion)

	hc := c.HTTP
	if hc == nil {
		hc = &http.Client{Timeout: defaultHTTPBudget}
	}
	resp, err := hc.Do(httpReq)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return Response{}, err
		}
		// Anything else at this level is "the endpoint is not there":
		// DNS, TCP, TLS, a proxy that ate it.
		return Response{}, fmt.Errorf("%w: %v", ErrOutage, err)
	}
	defer func() { _ = resp.Body.Close() }()

	// The reply is a few hundred bytes. Cap the read so a
	// misconfigured endpoint cannot stream a game server out of
	// memory.
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Response{}, fmt.Errorf("anthropic: read response: %w", err)
	}
	var out wireResponse
	if jerr := json.Unmarshal(payload, &out); jerr != nil && resp.StatusCode/100 == 2 {
		return Response{}, fmt.Errorf("anthropic: decode response: %w", jerr)
	}
	if resp.StatusCode/100 != 2 {
		e := &APIError{Status: resp.StatusCode}
		if out.Error != nil {
			e.Kind, e.Message = out.Error.Type, out.Error.Message
		} else {
			e.Message = strings.TrimSpace(string(payload))
		}
		return Response{}, e
	}

	var text strings.Builder
	for _, blk := range out.Content {
		if blk.Type == "text" {
			text.WriteString(blk.Text)
		}
	}
	return Response{
		Text:       text.String(),
		Model:      out.Model,
		StopReason: out.StopReason,
		Usage: Usage{
			InputTokens:      out.Usage.InputTokens,
			OutputTokens:     out.Usage.OutputTokens,
			CacheReadTokens:  out.Usage.CacheReadInputTokens,
			CacheWriteTokens: out.Usage.CacheCreationInputTokens,
		},
	}, nil
}
