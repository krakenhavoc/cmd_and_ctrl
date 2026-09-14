package model

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// openai_test.go pins the local-LLM transport's wire shape against an
// httptest server, the same way anthropic_test.go does and with the
// same honest limit: this proves the request this code BUILDS and the
// reply it can read, not that Ollama accepts either. The container
// has no network and no local model, so that is the whole of what can
// be checked here — and it is the half that regresses silently.

func serveOpenAI(t *testing.T, h http.HandlerFunc) *OpenAIClient {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &OpenAIClient{Endpoint: srv.URL, HTTP: srv.Client()}
}

func TestOpenAIRequestShape(t *testing.T) {
	var got map[string]any
	var raw string
	var headers http.Header
	var path string
	c := serveOpenAI(t, func(w http.ResponseWriter, r *http.Request) {
		headers = r.Header.Clone()
		path = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		raw = string(body)
		_ = json.Unmarshal(body, &got)
		_, _ = w.Write([]byte(`{"model":"llama3.1:8b","choices":[
			{"index":0,"message":{"role":"assistant","content":"{\"index\": 2}"},"finish_reason":"stop"}],
			"usage":{"prompt_tokens":900,"completion_tokens":11}}`))
	})
	c.APIKey = "local-key"

	resp, err := c.Complete(context.Background(), Request{
		Model:     "llama3.1:8b",
		System:    []Block{{Text: "primer"}, {Text: "decklist", Cache: true}},
		User:      "the board",
		MaxTokens: 128,
		// An Anthropic-only knob. It must not reach this endpoint:
		// several of these servers 400 on an unknown field.
		Effort: "low",
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}

	if path != "/v1/chat/completions" {
		t.Errorf("POSTed to %q, want /v1/chat/completions", path)
	}
	if headers.Get("authorization") != "Bearer local-key" {
		t.Errorf("authorization = %q", headers.Get("authorization"))
	}
	if got["model"] != "llama3.1:8b" {
		t.Errorf("model = %v", got["model"])
	}
	if got["max_tokens"] != float64(128) {
		t.Errorf("max_tokens = %v", got["max_tokens"])
	}
	for _, forbidden := range []string{"cache_control", "thinking", "output_config", "effort", "system\":"} {
		if strings.Contains(raw, forbidden) {
			t.Errorf("request carries %q, which this endpoint does not take:\n%s", forbidden, raw)
		}
	}

	// The one extra field this transport DOES send, and it sends
	// false: a hybrid-thinking model left on its default spends the
	// whole deadline thinking and answers nothing, which the funnel
	// scores as a timeout and covers with the heuristic's move.
	if got["think"] != false {
		t.Errorf("think = %v, want false — thinking must default OFF on this transport:\n%s", got["think"], raw)
	}

	// The system blocks are flattened into ONE system message, in
	// order, and the board travels as the user turn.
	msgs, _ := got["messages"].([]any)
	if len(msgs) != 2 {
		t.Fatalf("messages = %v, want a system and a user message", msgs)
	}
	sys, _ := msgs[0].(map[string]any)
	if sys["role"] != "system" {
		t.Errorf("first message role = %v", sys["role"])
	}
	if content, _ := sys["content"].(string); !strings.Contains(content, "primer") || !strings.Contains(content, "decklist") {
		t.Errorf("system message lost a block: %q", content)
	}
	user, _ := msgs[1].(map[string]any)
	if user["role"] != "user" || user["content"] != "the board" {
		t.Errorf("user message = %v", user)
	}

	if resp.Text != `{"index": 2}` {
		t.Errorf("Text = %q", resp.Text)
	}
	if resp.Model != "llama3.1:8b" || resp.StopReason != "stop" {
		t.Errorf("Response = %+v", resp)
	}
	if resp.Usage.InputTokens != 900 || resp.Usage.OutputTokens != 11 {
		t.Errorf("Usage = %+v", resp.Usage)
	}
	// The cache fields stay ZERO rather than being invented. This
	// endpoint has no prompt cache, and a fabricated number here
	// would corrupt the one measurement the funnel's cost argument
	// rests on.
	if resp.Usage.CacheReadTokens != 0 || resp.Usage.CacheWriteTokens != 0 {
		t.Errorf("cache tokens reported by an endpoint with no cache: %+v", resp.Usage)
	}
}

// A local server usually authenticates nobody. An empty key is a
// legal configuration here, unlike the Anthropic client.
func TestOpenAIWithoutAKeySendsNoAuthorization(t *testing.T) {
	var headers http.Header
	c := serveOpenAI(t, func(w http.ResponseWriter, r *http.Request) {
		headers = r.Header.Clone()
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"3"}}]}`))
	})
	resp, err := c.Complete(context.Background(), Request{Model: "m", User: "u"})
	if err != nil {
		t.Fatalf("Complete with no key: %v", err)
	}
	if _, ok := headers["Authorization"]; ok {
		t.Errorf("sent an Authorization header with no key configured: %q", headers.Get("authorization"))
	}
	if resp.Text != "3" {
		t.Errorf("Text = %q", resp.Text)
	}
}

// The four servers this targets are configured three different ways.
// Guessing correctly is worth more than insisting on one spelling.
func TestOpenAIEndpointCompletion(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"http://localhost:11434", "http://localhost:11434/v1/chat/completions"},
		{"http://localhost:11434/", "http://localhost:11434/v1/chat/completions"},
		{"http://localhost:1234/v1", "http://localhost:1234/v1/chat/completions"},
		{"http://localhost:8000/v1/", "http://localhost:8000/v1/chat/completions"},
		{"http://localhost:8080/v1/chat/completions", "http://localhost:8080/v1/chat/completions"},
	} {
		if got := (&OpenAIClient{Endpoint: tc.in}).URL(); got != tc.want {
			t.Errorf("URL(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// A non-2xx reply is typed, so a log can tell a 404 (the model id is
// not one this server serves — the commonest local misconfiguration)
// from a connection that never landed.
func TestOpenAISurfacesAPIErrors(t *testing.T) {
	c := serveOpenAI(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":{"type":"not_found","message":"model \"nope\" not found"}}`))
	})
	_, err := c.Complete(context.Background(), Request{Model: "nope", User: "u"})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *APIError", err)
	}
	if apiErr.Status != 404 || apiErr.Provider != "openai" {
		t.Errorf("APIError = %+v", apiErr)
	}
	if !strings.Contains(apiErr.Error(), "openai:") {
		t.Errorf("error message does not name the transport: %q", apiErr.Error())
	}
}

// An unreachable endpoint is ErrOutage, which on a local deployment
// is the likely error rather than an exotic one: the model server is
// simply not running yet.
func TestOpenAIUnreachableEndpointIsAnOutage(t *testing.T) {
	c := &OpenAIClient{Endpoint: "http://127.0.0.1:1", HTTP: &http.Client{Timeout: time.Second}}
	if _, err := c.Complete(context.Background(), Request{Model: "m", User: "u"}); !errors.Is(err, ErrOutage) {
		t.Errorf("err = %v, want ErrOutage", err)
	}
}

// ctx is the deadline that matters. A local model that thinks past it
// must return the context error — that is what the funnel counts as a
// timeout and what makes the "your model is too slow" warning fire
// rather than a generic failure.
func TestOpenAIRespectsTheContextDeadline(t *testing.T) {
	// The handler is released by the test rather than by the request
	// context, so that closing the server never waits on a timer.
	release := make(chan struct{})
	c := serveOpenAI(t, func(http.ResponseWriter, *http.Request) {
		<-release
	})
	t.Cleanup(func() { close(release) })
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := c.Complete(ctx, Request{Model: "m", User: "u"})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context.DeadlineExceeded", err)
	}
}

// A reply with no choices is not an answer. It must be an error
// rather than an empty string the parser would call malformed.
func TestOpenAIEmptyChoicesIsAnError(t *testing.T) {
	c := serveOpenAI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"model":"m","choices":[]}`))
	})
	if _, err := c.Complete(context.Background(), Request{Model: "m", User: "u"}); err == nil {
		t.Error("a reply with no choices was accepted")
	}
}

// A client with no endpoint is not a client. Nothing should dial.
func TestOpenAIWithoutAnEndpoint(t *testing.T) {
	var c *OpenAIClient
	if _, err := c.Complete(context.Background(), Request{}); !errors.Is(err, ErrNoClient) {
		t.Errorf("nil client: err = %v, want ErrNoClient", err)
	}
	if _, err := (&OpenAIClient{}).Complete(context.Background(), Request{}); !errors.Is(err, ErrNoClient) {
		t.Errorf("empty endpoint: err = %v, want ErrNoClient", err)
	}
}

// The transport is enabled by one variable and nothing else, so that
// "is the bot using my local model" has a single answer.
func TestNewOpenAIClientFromEnv(t *testing.T) {
	t.Setenv(EnvOpenAIEndpoint, "")
	if c := NewOpenAIClient(); c != nil {
		t.Fatalf("built a client with no endpoint configured: %+v", c)
	}
	t.Setenv(EnvOpenAIEndpoint, "http://localhost:11434")
	t.Setenv(EnvOpenAIKey, "k")
	c := NewOpenAIClient()
	if c == nil {
		t.Fatal("no client built although the endpoint is configured")
	}
	if c.URL() != "http://localhost:11434/v1/chat/completions" || c.APIKey != "k" {
		t.Errorf("client = %+v", c)
	}
	if c.Name() != "openai-compatible" {
		t.Errorf("Name() = %q", c.Name())
	}
}

// The funnel's own accounting of a timeout: a call that runs out of
// budget is recorded as a TIMEOUT, not folded into the generic
// failure count. On a self-hosted model this is the number that says
// "your `assisted` seat is playing the heuristic on every window".
func TestFunnelCountsTimeoutsSeparately(t *testing.T) {
	p := testPolicy(t, NeverAnswers(), &stubB{index: 1, reason: "the heuristic's pick"}, func(c *Config) {
		c.Reserve = 50 * time.Millisecond
		c.MinBudget = 10 * time.Millisecond
	})
	if d := decide(t, p, castWindow(), 400*time.Millisecond); d.Index != 1 {
		t.Fatalf("index = %d, want Layer B's 1", d.Index)
	}
	st := p.Stats()
	if st.ModelTimeouts != 1 {
		t.Errorf("ModelTimeouts = %d, want 1", st.ModelTimeouts)
	}
	if st.ByFallback[FallbackError] != 1 {
		t.Errorf("ByFallback[error] = %d, want 1 — a timeout is still a call that failed", st.ByFallback[FallbackError])
	}
	recs := p.Records()
	if len(recs) != 1 || !recs[0].TimedOut {
		t.Errorf("record = %+v, want TimedOut", recs)
	}
}

// Thinking is off by default and back on when the caller asks for it
// by name, so Request.Thinking means something on this transport
// rather than being quietly dropped.
func TestOpenAIThinkingSwitch(t *testing.T) {
	for _, tc := range []struct {
		thinking string
		omit     bool
		want     any // nil means "the field must be absent"
	}{
		{thinking: "", want: false},
		{thinking: "disabled", want: false},
		{thinking: "adaptive", want: true},
		{thinking: "adaptive", omit: true, want: nil},
	} {
		var got map[string]any
		c := serveOpenAI(t, func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			got = nil
			_ = json.Unmarshal(body, &got)
			_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"0"}}]}`))
		})
		c.OmitThink = tc.omit
		if _, err := c.Complete(context.Background(), Request{Model: "m", User: "u", Thinking: tc.thinking}); err != nil {
			t.Fatalf("thinking=%q omit=%v: %v", tc.thinking, tc.omit, err)
		}
		v, present := got["think"]
		switch {
		case tc.want == nil && present:
			t.Errorf("thinking=%q omit=%v: think was sent (%v) although the field is suppressed", tc.thinking, tc.omit, v)
		case tc.want != nil && !present:
			t.Errorf("thinking=%q: think was not sent at all, so a thinking model stays on", tc.thinking)
		case tc.want != nil && v != tc.want:
			t.Errorf("thinking=%q: think = %v, want %v", tc.thinking, v, tc.want)
		}
	}
}

// "1" means "the Ollama on this machine". Anything else is a URL, so
// a model on another box on the LAN is one variable away and no IP is
// compiled in.
func TestNewOpenAIClientShorthandEndpoint(t *testing.T) {
	t.Setenv(EnvOpenAIKey, "")
	t.Setenv(EnvOpenAISendThink, "")
	for _, in := range []string{"1", "true", "default", "ollama"} {
		t.Setenv(EnvOpenAIEndpoint, in)
		c := NewOpenAIClient()
		if c == nil || c.URL() != "http://localhost:11434/v1/chat/completions" {
			t.Errorf("%q built %+v", in, c)
		}
	}
	t.Setenv(EnvOpenAIEndpoint, "http://192.0.2.10:11434")
	if got := NewOpenAIClient().URL(); got != "http://192.0.2.10:11434/v1/chat/completions" {
		t.Errorf("URL = %q", got)
	}
	// The escape hatch, for a server strict enough to reject the
	// unknown field.
	t.Setenv(EnvOpenAISendThink, "0")
	if c := NewOpenAIClient(); !c.OmitThink {
		t.Error("CMDCTRL_OPENAI_SEND_THINK=0 did not stop the think field")
	}
	t.Setenv(EnvOpenAISendThink, "1")
	if c := NewOpenAIClient(); c.OmitThink {
		t.Error("CMDCTRL_OPENAI_SEND_THINK=1 suppressed the think field")
	}
}
