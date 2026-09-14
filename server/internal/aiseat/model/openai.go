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

// openai.go is the second transport, and it exists because the first
// one costs money. It speaks POST /v1/chat/completions — the shape
// Ollama, LM Studio, llama.cpp's server and vLLM all expose — so a
// deployment can point the bot at a model running on its own hardware
// and never touch a hosted API.
//
// anthropic.go's doc says the swap point is "implement Client with
// the SDK and delete this file". This is that seam used in the other
// direction: a second implementation of a two-method interface, under
// the same constraints as the first — raw net/http (server/go.mod has
// three direct dependencies and pins go 1.22 on purpose), no
// streaming, no tools, no retries, and ctx is the deadline that
// matters.
//
// # What is deliberately NOT sent
//
// Prompt caching is Anthropic-specific. There is no
// OpenAI-compatible equivalent of a cache breakpoint, so Block.Cache
// is dropped rather than translated into something that looks like
// it: the system blocks are flattened into one system message, and
// the cache fields of Usage come back ZERO rather than guessed. A
// local deployment's metrics then say, truthfully, that it paid no
// cache reads — which is also why the cost arithmetic in ADR 0033 §5
// does not apply to it (there is no per-token cost to account for).
//
// Nothing else Anthropic-shaped travels either: no cache_control, no
// thinking, no output_config.effort. Several of these servers are
// strict about unknown fields and answer 400, and a model that has
// never heard of "effort" loses nothing by not being asked for it.
//
// # The deadline is the real hazard
//
// ADR 0033 §10 gives `assisted` a 2s hard deadline, sized for a
// hosted cheap model. A 7B model on consumer hardware will blow that
// routinely, and the funnel's designed response — fall back to Layer
// B — would then fire on every single window: heuristic play wearing
// the "assisted" label, which is exactly the silent downgrade
// aiseat/tier.go forbids. So the deadline is configurable per
// deployment (CMDCTRL_BOT_MAX_THINK, wired in main.go), the local
// transport gets a larger default, and a call that runs out of time
// is logged as a TIMEOUT rather than folded into the generic failure
// count — see Stats.ModelTimeouts.

const (
	// openAIDefaultPath is appended to an endpoint that names only a
	// host or an API root.
	openAIChatPath = "/chat/completions"
	openAIAPIRoot  = "/v1"
	// openAIDefaultBudget is this transport's own HTTP timeout when
	// the caller supplies no deadline. Generous next to the
	// Anthropic client's because the machine answering is usually
	// under the same desk.
	openAIDefaultBudget = 120 * time.Second
)

// OpenAIClient calls an OpenAI-compatible /v1/chat/completions
// endpoint. It is the local-LLM transport, and it is equally happy
// talking to a hosted provider that speaks the same shape.
type OpenAIClient struct {
	// Endpoint is the chat-completions URL. A bare host
	// ("http://localhost:11434") or an API root
	// ("http://localhost:1234/v1") is completed for you. Required.
	Endpoint string
	// APIKey is sent as `Authorization: Bearer`. EMPTY IS LEGAL and
	// is the common case: a local server usually authenticates
	// nobody. This is the one behavioural difference from the
	// Anthropic client, where an empty key means "no transport".
	APIKey string
	// Label names the transport in logs and in the startup banner.
	// Empty reports "openai-compatible".
	Label string
	// HTTP is the client used for the call. Nil builds one with a
	// conservative timeout; the per-call deadline that actually
	// matters comes from ctx.
	HTTP *http.Client
}

// Env vars this transport reads. Named for the wire shape rather
// than for any one server, because the whole point is that four of
// them speak it.
const (
	// EnvOpenAIEndpoint turns this transport ON. Unset means the
	// server has no local model and NewOpenAIClient returns nil.
	EnvOpenAIEndpoint = "CMDCTRL_OPENAI_ENDPOINT"
	// EnvOpenAIKey is optional — see OpenAIClient.APIKey.
	EnvOpenAIKey = "CMDCTRL_OPENAI_API_KEY"
)

// NewOpenAIClient builds the transport from the environment, and
// returns nil when CMDCTRL_OPENAI_ENDPOINT is unset. A nil Client is
// not an error anywhere in this package: it is how a deployment
// without a local model falls through to the Anthropic client, or to
// no model at all.
func NewOpenAIClient() *OpenAIClient {
	endpoint := strings.TrimSpace(os.Getenv(EnvOpenAIEndpoint))
	if endpoint == "" {
		return nil
	}
	return &OpenAIClient{
		Endpoint: endpoint,
		APIKey:   strings.TrimSpace(os.Getenv(EnvOpenAIKey)),
	}
}

// Name identifies the transport.
func (c *OpenAIClient) Name() string {
	if c != nil && c.Label != "" {
		return c.Label
	}
	return "openai-compatible"
}

// URL is the fully-qualified chat-completions endpoint this client
// will POST to. Exported because the boot log should say exactly
// where the bot is dialling — a typo'd port is otherwise a per-window
// warning and nothing else.
func (c *OpenAIClient) URL() string {
	return openAIChatURL(c.Endpoint)
}

// openAIChatURL completes a configured endpoint. The four servers
// this targets are configured by their users in three different ways
// — bare host, API root, or the full path — and guessing correctly is
// worth more than insisting on one of them.
func openAIChatURL(endpoint string) string {
	e := strings.TrimRight(strings.TrimSpace(endpoint), "/")
	switch {
	case e == "":
		return ""
	case strings.HasSuffix(e, openAIChatPath):
		return e
	case strings.HasSuffix(e, openAIAPIRoot):
		return e + openAIChatPath
	default:
		return e + openAIAPIRoot + openAIChatPath
	}
}

// --- wire types ----------------------------------------------------

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRequest struct {
	Model     string          `json:"model"`
	Messages  []openAIMessage `json:"messages"`
	MaxTokens int             `json:"max_tokens"`
}

type openAIChoice struct {
	Index        int           `json:"index"`
	Message      openAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

type openAIResponse struct {
	Model   string         `json:"model"`
	Choices []openAIChoice `json:"choices"`
	Usage   openAIUsage    `json:"usage"`
	Error   *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// Complete makes one chat-completions call.
func (c *OpenAIClient) Complete(ctx context.Context, req Request) (Response, error) {
	if c == nil || strings.TrimSpace(c.Endpoint) == "" {
		return Response{}, ErrNoClient
	}
	body := openAIRequest{
		Model:     req.Model,
		MaxTokens: req.MaxTokens,
	}
	if body.MaxTokens <= 0 {
		body.MaxTokens = 256
	}
	// The static blocks become ONE system message. The cache
	// breakpoint is dropped, not translated: see the file comment.
	if sys := flattenBlocks(req.System); sys != "" {
		body.Messages = append(body.Messages, openAIMessage{Role: "system", Content: sys})
	}
	body.Messages = append(body.Messages, openAIMessage{Role: "user", Content: req.User})

	raw, err := json.Marshal(body)
	if err != nil {
		return Response{}, fmt.Errorf("openai: marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL(), bytes.NewReader(raw))
	if err != nil {
		return Response{}, fmt.Errorf("openai: build request: %w", err)
	}
	httpReq.Header.Set("content-type", "application/json")
	if c.APIKey != "" {
		httpReq.Header.Set("authorization", "Bearer "+c.APIKey)
	}

	hc := c.HTTP
	if hc == nil {
		hc = &http.Client{Timeout: openAIDefaultBudget}
	}
	resp, err := hc.Do(httpReq)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return Response{}, err
		}
		// The local case makes this the LIKELY error rather than an
		// exotic one: the model server is simply not running yet.
		return Response{}, fmt.Errorf("%w: %v", ErrOutage, err)
	}
	defer func() { _ = resp.Body.Close() }()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Response{}, fmt.Errorf("openai: read response: %w", err)
	}
	var out openAIResponse
	if jerr := json.Unmarshal(payload, &out); jerr != nil && resp.StatusCode/100 == 2 {
		return Response{}, fmt.Errorf("openai: decode response: %w", jerr)
	}
	if resp.StatusCode/100 != 2 {
		e := &APIError{Status: resp.StatusCode, Provider: "openai"}
		if out.Error != nil {
			e.Kind, e.Message = out.Error.Type, out.Error.Message
		} else {
			e.Message = strings.TrimSpace(string(payload))
		}
		return Response{}, e
	}
	if len(out.Choices) == 0 {
		return Response{}, fmt.Errorf("openai: reply had no choices")
	}
	return Response{
		Text:       out.Choices[0].Message.Content,
		Model:      out.Model,
		StopReason: out.Choices[0].FinishReason,
		Usage: Usage{
			InputTokens:  out.Usage.PromptTokens,
			OutputTokens: out.Usage.CompletionTokens,
			// Cache reads and writes stay ZERO. This endpoint has no
			// prompt cache to report, and a fabricated number here
			// would corrupt the one measurement the funnel's cost
			// argument rests on.
		},
	}, nil
}

// flattenBlocks joins the system blocks into one message. Order is
// preserved; the Cache flag is ignored, because there is nothing on
// this wire to place a breakpoint with.
func flattenBlocks(blocks []Block) string {
	parts := make([]string, 0, len(blocks))
	for _, b := range blocks {
		if s := strings.TrimSpace(b.Text); s != "" {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, "\n\n")
}
