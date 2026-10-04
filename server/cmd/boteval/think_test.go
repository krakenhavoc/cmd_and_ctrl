package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/suite"
)

// think_test.go covers the thinking experiment's flags (#2196) end to
// end where it matters: the request that reaches the endpoint, and the
// deadline it is given.

func clearThinkEnv(t *testing.T) {
	t.Helper()
	t.Setenv(model.EnvBotThink, "")
	t.Setenv(model.EnvBotMaxTokens, "")
	t.Setenv("CMDCTRL_BOT_MAX_THINK", "")
}

// suiteRequest runs one position through `suite run` with args against
// a stub transport and returns the request body and the deadline the
// transport saw.
func suiteRequest(t *testing.T, args ...string) (map[string]any, time.Duration, suite.Report) {
	t.Helper()
	t.Setenv("CMDCTRL_OPENAI_ENDPOINT", "http://review.invalid")
	t.Setenv("CMDCTRL_BOT_MODEL", "review-model")
	var (
		body   map[string]any
		budget time.Duration
	)
	previous := http.DefaultTransport
	http.DefaultTransport = deadlineCaptureTransport(func(r *http.Request) (*http.Response, error) {
		deadline, _ := r.Context().Deadline()
		budget = time.Until(deadline)
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		return &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
			Body: io.NopCloser(strings.NewReader(
				`{"choices":[{"message":{"content":"0","reasoning":"thought about it"},"finish_reason":"stop"}],"usage":{"completion_tokens":900}}`)),
		}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = previous })

	opts, err := parseSuiteRun(append([]string{"--policy", "assisted"}, args...))
	if err != nil {
		t.Fatal(err)
	}
	policy, note, err := buildSuitePolicy(os.Stdout, opts)
	if err != nil {
		t.Fatal(err)
	}
	pos, err := suite.LoadFile("../../internal/aiseat/suite/testdata/positions/attack-an-empty-board.json")
	if err != nil {
		t.Fatal(err)
	}
	rep := suite.Run(context.Background(), []suite.Position{pos}, policy, suite.RunOptions{MaxThink: opts.MaxThink})
	rep.Note = note
	if body == nil {
		t.Fatal("no request reached the transport")
	}
	return body, budget, rep
}

func TestSuiteThinkSendsNoThinkingOffFieldsAndWaitsLonger(t *testing.T) {
	clearThinkEnv(t)
	body, budget, rep := suiteRequest(t, "--think")
	if v, ok := body["reasoning_effort"]; ok {
		t.Errorf("reasoning_effort = %v was sent with --think", v)
	}
	if body["think"] != true {
		t.Errorf("think = %v, want true", body["think"])
	}
	if body["max_tokens"] != float64(model.DefaultThinkingMaxTokens) {
		t.Errorf("max_tokens = %v, want %d", body["max_tokens"], model.DefaultThinkingMaxTokens)
	}
	if budget < model.DefaultThinkingMaxThink-30*time.Second {
		t.Errorf("a --think run gave the call %v; it needs about %v", budget, model.DefaultThinkingMaxThink)
	}
	if !strings.Contains(rep.Note, "thinking on, max tokens 8000") {
		t.Errorf("report note %q does not record the experiment", rep.Note)
	}
	if rep.ReasoningReplies != 1 || rep.ReasoningCharsP50 != len("thought about it") {
		t.Errorf("reasoning = %d replies p50 %d", rep.ReasoningReplies, rep.ReasoningCharsP50)
	}
	if !strings.Contains(rep.Markdown(), "thinking on") {
		t.Error("the Markdown block does not say thinking was on")
	}
}

func TestSuiteDefaultStillSuppressesThinking(t *testing.T) {
	clearThinkEnv(t)
	body, _, rep := suiteRequest(t)
	if body["think"] != false || body["reasoning_effort"] != "none" {
		t.Errorf("think = %v, reasoning_effort = %v; the default must keep thinking off", body["think"], body["reasoning_effort"])
	}
	if mt := body["max_tokens"].(float64); mt > 256 {
		t.Errorf("max_tokens = %v, want the shipped 128/256", mt)
	}
	if !strings.Contains(rep.Note, "thinking off") {
		t.Errorf("report note %q", rep.Note)
	}
}

func TestSuiteThinkComesFromTheEnvironmentToo(t *testing.T) {
	clearThinkEnv(t)
	t.Setenv(model.EnvBotThink, "1")
	t.Setenv(model.EnvBotMaxTokens, "3000")
	body, _, _ := suiteRequest(t, "--max-think", "45s")
	if body["think"] != true || body["max_tokens"] != float64(3000) {
		t.Errorf("think = %v max_tokens = %v, want true / 3000", body["think"], body["max_tokens"])
	}
}

func TestSuiteMaxTokensWithoutThink(t *testing.T) {
	clearThinkEnv(t)
	body, _, _ := suiteRequest(t, "--max-tokens", "512")
	if body["reasoning_effort"] != "none" || body["max_tokens"] != float64(512) {
		t.Errorf("reasoning_effort = %v max_tokens = %v, want none / 512", body["reasoning_effort"], body["max_tokens"])
	}
}

func TestParseSuiteRunRefusesABadThinkVariable(t *testing.T) {
	clearThinkEnv(t)
	t.Setenv(model.EnvBotThink, "perhaps")
	if _, err := parseSuiteRun([]string{"--policy", "assisted"}); err == nil {
		t.Fatal("a malformed CMDCTRL_BOT_THINK was accepted; the run would report the baseline under the experiment's name")
	}
}

func TestParseArenaFlagsThink(t *testing.T) {
	clearThinkEnv(t)
	t.Setenv("CMDCTRL_OPENAI_ENDPOINT", "http://localhost:11434")
	a, err := parseArenaFlags([]string{"--seats", "assisted,heuristic", "--think"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if !a.think || a.maxTokens != model.DefaultThinkingMaxTokens {
		t.Errorf("think %v max tokens %d", a.think, a.maxTokens)
	}
	if a.maxThink != model.DefaultThinkingMaxThink || !a.thinkDefaulted {
		t.Errorf("max think %v (defaulted %v), want %v", a.maxThink, a.thinkDefaulted, model.DefaultThinkingMaxThink)
	}
	cfg := a.config(nil, nil, nil, "", nil)
	if !cfg.Think || cfg.MaxTokens != model.DefaultThinkingMaxTokens {
		t.Errorf("botarena config think %v max tokens %d", cfg.Think, cfg.MaxTokens)
	}

	// An explicit deadline and budget win.
	a, err = parseArenaFlags([]string{"--seats", "assisted,heuristic", "--think", "--max-tokens", "4000", "--max-think", "60s"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if a.maxTokens != 4000 || a.maxThink != 60*time.Second || a.thinkDefaulted {
		t.Errorf("max tokens %d max think %v defaulted %v", a.maxTokens, a.maxThink, a.thinkDefaulted)
	}

	// And no --think leaves the local default alone.
	a, err = parseArenaFlags([]string{"--seats", "assisted,heuristic"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if a.think || a.maxTokens != 0 || a.maxThink != localDefaultMaxThink {
		t.Errorf("default run: think %v max tokens %d max think %v", a.think, a.maxTokens, a.maxThink)
	}
}

func TestThinkingOnVerdict(t *testing.T) {
	base := probeResult{Think: true, MaxTokens: 8000, PromptTokens: 4800, SystemBytes: 15000, Moves: 5, CompletionTokens: 1500}
	answered := base
	answered.Reasoning, answered.Reply, answered.FinishReason = "hmm", `{"index":0}`, "stop"
	if v := thinkingVerdict(answered); !strings.Contains(v, "on, as asked") {
		t.Errorf("answered: %s", v)
	}
	ranOut := base
	ranOut.Reasoning, ranOut.FinishReason = "hmm", "length"
	if v := thinkingVerdict(ranOut); !strings.Contains(v, "RAN OUT OF BUDGET") {
		t.Errorf("ran out: %s", v)
	}
	silent := base
	silent.Reply, silent.FinishReason = `{"index":0}`, "stop"
	if v := thinkingVerdict(silent); !strings.Contains(v, "no reasoning came back") {
		t.Errorf("silent: %s", v)
	}
}
