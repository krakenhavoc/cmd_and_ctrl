package model

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// anthropic_test.go pins the wire shape against an httptest server.
// It proves the request this code BUILDS, not that the API accepts
// it — that distinction is the honest limit of what can be tested
// without a key, and it is worth being explicit about: a schema
// change at the provider would pass every test here and fail in
// production. What these tests do catch is the class of bug that is
// otherwise invisible until the bill arrives, chiefly a cache
// breakpoint that stops being sent.

func serve(t *testing.T, h http.HandlerFunc) (*AnthropicClient, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &AnthropicClient{APIKey: "test-key", Endpoint: srv.URL, HTTP: srv.Client()}, srv
}

func TestAnthropicRequestShape(t *testing.T) {
	var got map[string]any
	var headers http.Header
	c, _ := serve(t, func(w http.ResponseWriter, r *http.Request) {
		headers = r.Header.Clone()
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		_, _ = w.Write([]byte(`{"model":"m","stop_reason":"end_turn",
			"content":[{"type":"text","text":"{\"index\": 2}"}],
			"usage":{"input_tokens":900,"output_tokens":11,"cache_creation_input_tokens":800,"cache_read_input_tokens":0}}`))
	})

	resp, err := c.Complete(context.Background(), Request{
		Model:     "claude-haiku-4-5",
		System:    []Block{{Text: "primer"}, {Text: "decklist", Cache: true}},
		User:      "the board",
		MaxTokens: 128,
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}

	if headers.Get("x-api-key") != "test-key" {
		t.Errorf("x-api-key = %q", headers.Get("x-api-key"))
	}
	if headers.Get("anthropic-version") == "" {
		t.Error("anthropic-version header is missing; the API requires it")
	}
	if got["model"] != "claude-haiku-4-5" {
		t.Errorf("model = %v", got["model"])
	}
	if got["max_tokens"] != float64(128) {
		t.Errorf("max_tokens = %v", got["max_tokens"])
	}
	// The cache breakpoint is the whole cost argument; it has to be on
	// the LAST system block and nowhere else.
	sys, _ := got["system"].([]any)
	if len(sys) != 2 {
		t.Fatalf("system blocks = %d, want 2", len(sys))
	}
	if _, ok := sys[0].(map[string]any)["cache_control"]; ok {
		t.Error("the first system block carries a breakpoint; only the last should")
	}
	cc, ok := sys[1].(map[string]any)["cache_control"].(map[string]any)
	if !ok || cc["type"] != "ephemeral" {
		t.Errorf("the last system block has no ephemeral cache_control: %v", sys[1])
	}
	msgs, _ := got["messages"].([]any)
	if len(msgs) != 1 || msgs[0].(map[string]any)["role"] != "user" {
		t.Errorf("messages = %v", msgs)
	}
	// Empty Effort and Thinking must be OMITTED rather than sent
	// empty: some models reject the fields outright.
	if _, ok := got["output_config"]; ok {
		t.Error("output_config was sent for an empty Effort")
	}
	if _, ok := got["thinking"]; ok {
		t.Error("thinking was sent for an empty Thinking")
	}

	if resp.Text != `{"index": 2}` {
		t.Errorf("text = %q", resp.Text)
	}
	if resp.Usage.InputTokens != 900 || resp.Usage.CacheWriteTokens != 800 {
		t.Errorf("usage = %+v", resp.Usage)
	}
	if resp.StopReason != "end_turn" {
		t.Errorf("stop reason = %q", resp.StopReason)
	}
}

func TestAnthropicSendsEffortAndThinkingWhenSet(t *testing.T) {
	var got map[string]any
	c, _ := serve(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		_, _ = w.Write([]byte(`{"content":[{"type":"text","text":"0"}]}`))
	})
	if _, err := c.Complete(context.Background(), Request{
		Model: "claude-opus-5", User: "x", Effort: "low", Thinking: "adaptive",
	}); err != nil {
		t.Fatal(err)
	}
	oc, _ := got["output_config"].(map[string]any)
	if oc["effort"] != "low" {
		t.Errorf("output_config = %v", got["output_config"])
	}
	th, _ := got["thinking"].(map[string]any)
	if th["type"] != "adaptive" {
		t.Errorf("thinking = %v", got["thinking"])
	}
}

func TestAnthropicConcatenatesTextBlocks(t *testing.T) {
	c, _ := serve(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"content":[
			{"type":"thinking","text":"ignored"},
			{"type":"text","text":"{\"index\":"},
			{"type":"text","text":" 4}"}]}`))
	})
	resp, err := c.Complete(context.Background(), Request{Model: "m", User: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Text != `{"index": 4}` {
		t.Errorf("text = %q", resp.Text)
	}
}

func TestAnthropicSurfacesAPIErrors(t *testing.T) {
	c, _ := serve(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"type":"rate_limit_error","message":"slow down"}}`))
	})
	_, err := c.Complete(context.Background(), Request{Model: "m", User: "x"})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *APIError", err)
	}
	if apiErr.Status != 429 || apiErr.Kind != "rate_limit_error" {
		t.Errorf("err = %+v", apiErr)
	}
}

// A 401 has to be distinguishable from an outage: one is a deploy
// problem somebody must fix, the other is weather.
func TestAnthropicDistinguishesAuthFailureFromOutage(t *testing.T) {
	c, _ := serve(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"type":"authentication_error","message":"invalid x-api-key"}}`))
	})
	_, err := c.Complete(context.Background(), Request{Model: "m", User: "x"})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != 401 {
		t.Fatalf("err = %v, want a 401 APIError", err)
	}
	if errors.Is(err, ErrOutage) {
		t.Error("a bad key was reported as an outage")
	}
}

func TestAnthropicUnreachableEndpointIsAnOutage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close() // nothing is listening any more
	c := &AnthropicClient{APIKey: "k", Endpoint: url, HTTP: &http.Client{Timeout: time.Second}}
	_, err := c.Complete(context.Background(), Request{Model: "m", User: "x"})
	if !errors.Is(err, ErrOutage) {
		t.Fatalf("err = %v, want ErrOutage", err)
	}
}

// The policy hands the transport a budget and expects it back on
// time. A transport that ignored ctx would blow MaxThink and the
// runner would force-pass.
func TestAnthropicRespectsTheDeadline(t *testing.T) {
	// The handler hangs until either the client gives up or the test
	// releases it. The release channel is not belt-and-braces: a
	// server whose only exit is the request context can outlive the
	// client's cancellation, and httptest.Server.Close waits for
	// every handler to return — which turns a passing test into a
	// ten-minute hang under -race.
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))
	defer srv.Close()
	defer close(release)
	c := &AnthropicClient{APIKey: "k", Endpoint: srv.URL, HTTP: srv.Client()}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	started := time.Now()
	_, err := c.Complete(ctx, Request{Model: "m", User: "x"})
	if err == nil {
		t.Fatal("a hung endpoint returned no error")
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Errorf("Complete took %v against a 150ms deadline", elapsed)
	}
	if errors.Is(err, ErrOutage) {
		t.Errorf("a deadline was reported as an outage: %v", err)
	}
}

func TestAnthropicWithoutAKeyIsNotAClient(t *testing.T) {
	c := &AnthropicClient{}
	if _, err := c.Complete(context.Background(), Request{Model: "m"}); !errors.Is(err, ErrNoClient) {
		t.Errorf("err = %v, want ErrNoClient", err)
	}
	var nilClient *AnthropicClient
	if _, err := nilClient.Complete(context.Background(), Request{Model: "m"}); !errors.Is(err, ErrNoClient) {
		t.Errorf("nil client err = %v, want ErrNoClient", err)
	}
}

func TestNewAnthropicClientReadsTheEnvironment(t *testing.T) {
	t.Setenv("CMDCTRL_ANTHROPIC_API_KEY", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	if c := NewAnthropicClient(); c != nil {
		t.Fatal("a client was built with no key; a keyless deployment must get nil and run on Layer A + B")
	}
	t.Setenv("ANTHROPIC_API_KEY", "from-anthropic-env")
	if c := NewAnthropicClient(); c == nil || c.APIKey != "from-anthropic-env" {
		t.Fatalf("client = %+v", c)
	}
	t.Setenv("CMDCTRL_ANTHROPIC_API_KEY", "project-specific")
	if c := NewAnthropicClient(); c == nil || c.APIKey != "project-specific" {
		t.Errorf("the project-specific key did not win: %+v", c)
	}
	t.Setenv("CMDCTRL_ANTHROPIC_ENDPOINT", "https://example.invalid/v1/messages")
	if c := NewAnthropicClient(); c == nil || c.Endpoint != "https://example.invalid/v1/messages" {
		t.Errorf("endpoint override ignored: %+v", c)
	}
}

// The whole funnel over a real HTTP round trip, with the transport in
// place of a fake. Still not a model — but it is the only test that
// runs prompt assembly, the wire encoding, the decode and the index
// validation as one path.
func TestTheFunnelOverHTTP(t *testing.T) {
	c, _ := serve(t, func(w http.ResponseWriter, r *http.Request) {
		var req wireRequest
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("the server could not decode the request: %v", err)
		}
		if len(req.Messages) != 1 || req.Messages[0].Content == "" {
			t.Errorf("no board reached the endpoint: %+v", req.Messages)
		}
		_, _ = w.Write([]byte(`{"model":"claude-haiku-4-5","content":[{"type":"text","text":"{\"index\": 1, \"why\": \"burn the leader\"}"}],
			"usage":{"input_tokens":1200,"output_tokens":9,"cache_read_input_tokens":1100}}`))
	})
	p := testPolicy(t, c, &stubB{index: 2, reason: "b"}, func(cfg *Config) { cfg.Deck = testDeck() })
	d := decide(t, p, castWindow(), 2*time.Second)
	if d.Index != 1 {
		t.Fatalf("index = %d, want the endpoint's 1", d.Index)
	}
	st := p.Stats()
	if st.ByLayer[LayerC] != 1 || st.ModelCalls != 1 {
		t.Errorf("stats = %+v", st)
	}
	if st.Usage.CacheReadTokens != 1100 {
		t.Errorf("cache read tokens were not recorded: %+v", st.Usage)
	}
}
