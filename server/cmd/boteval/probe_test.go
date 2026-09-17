package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
)

// probe_test.go pins the two verdicts. They are the whole output of
// the subcommand — everything above them is numbers, and the verdict
// line is what somebody acts on — so a verdict that fires on the
// wrong evidence is worse than no probe at all.
//
// The fixtures are the SHAPES measured on the real host (Ollama
// 0.34.0, qwen3:14b) rather than invented ones: with `think:false`
// alone on /v1/chat/completions the reply came back with empty
// content, finish_reason "length", and the whole thinking trace in
// message.reasoning. #839 fixed the sending side; this pins the
// reading side, so a server that honours neither switch is named
// rather than silently played around.

func TestThinkingVerdictFiresOnTheMeasuredQwenShape(t *testing.T) {
	cases := []struct {
		name string
		res  probeResult
		want string
	}{
		{
			name: "reasoning field present",
			res: probeResult{
				PromptTokens: 3000, Moves: 8, FinishReason: "length",
				Reasoning: "Okay, let me think about which move is best…",
			},
			want: "THINKING: not suppressed",
		},
		{
			name: "length with an empty reply and no reasoning field",
			res:  probeResult{PromptTokens: 3000, Moves: 8, FinishReason: "length", Reply: ""},
			want: "THINKING: not suppressed",
		},
		{
			name: "a clean answer",
			res: probeResult{
				PromptTokens: 3000, Moves: 8, FinishReason: "stop",
				Reply: `{"index": 3, "why": "removal"}`, ParsedIndex: 3,
			},
			want: "THINKING: suppressed",
		},
		{
			name: "length but the answer still arrived",
			res: probeResult{
				PromptTokens: 3000, Moves: 8, FinishReason: "length",
				Reply: `{"index": 3}`, ParsedIndex: 3,
			},
			want: "THINKING: suppressed",
		},
		{
			name: "the call failed",
			res:  probeResult{CallErr: errors.New("connection refused")},
			want: "THINKING: unknown",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := thinkingVerdict(tc.res); !strings.HasPrefix(got, tc.want) {
				t.Errorf("thinkingVerdict = %q, want it to start %q", got, tc.want)
			}
		})
	}
}

func TestTruncationVerdict(t *testing.T) {
	// 9000 bytes of prompt is ~3000 tokens by the client-side
	// estimate; a server that counted 1200 of them dropped the front.
	cases := []struct {
		name string
		res  probeResult
		want string
	}{
		{
			name: "server counted far fewer tokens than were sent",
			res: probeResult{
				SystemBytes: 8000, UserBytes: 1000, PromptTokens: 1200,
				Moves: 8, Reply: `{"index": 1}`, ParsedIndex: 1,
			},
			want: "TRUNCATION: likely",
		},
		{
			name: "token count fine, reply ignores the format",
			res: probeResult{
				SystemBytes: 8000, UserBytes: 1000, PromptTokens: 2900,
				Moves: 8, Reply: "I would cast Lightning Bolt.", ParseErr: errors.New("no index"),
			},
			want: "TRUNCATION: likely",
		},
		{
			name: "token count fine, in-range answer",
			res: probeResult{
				SystemBytes: 8000, UserBytes: 1000, PromptTokens: 2900,
				Moves: 8, Reply: `{"index": 1}`, ParsedIndex: 1,
			},
			want: "TRUNCATION: not detected",
		},
		{
			name: "an in-shape reply naming a move that does not exist is still not truncation evidence we can read",
			res: probeResult{
				SystemBytes: 8000, UserBytes: 1000, PromptTokens: 2900,
				Moves: 8, Reply: `{"index": 99}`, ParsedIndex: 99,
			},
			want: "TRUNCATION: likely",
		},
		{
			name: "endpoint reported no usage",
			res:  probeResult{SystemBytes: 8000, UserBytes: 1000, Moves: 8, ParsedIndex: 1},
			want: "TRUNCATION: unknown",
		},
		{
			name: "the call failed",
			res:  probeResult{CallErr: errors.New("connection refused")},
			want: "TRUNCATION: unknown",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := truncationVerdict(tc.res); !strings.HasPrefix(got, tc.want) {
				t.Errorf("truncationVerdict = %q, want it to start %q", got, tc.want)
			}
		})
	}
}

// The transport half: the probe's evidence comes off the wire through
// the ordinary OpenAI client, so the fields it reads have to be the
// fields a real server sends. This is the shape the dev box actually
// returns.
func TestProbeReadsTheQwenReplyOffTheWire(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(`{
			"model": "qwen3:14b",
			"choices": [{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": "",
					"reasoning": "Okay, the opponent has a 5/5 and I am at 12 life, so…"
				},
				"finish_reason": "length"
			}],
			"usage": {
				"prompt_tokens": 6231,
				"completion_tokens": 128,
				"prompt_tokens_details": {"cached_tokens": 4096}
			}
		}`))
	}))
	defer srv.Close()

	t.Setenv(model.EnvOpenAIEndpoint, srv.URL)
	client, url := buildClient("")
	if client == nil {
		t.Fatal("no client built from the endpoint")
	}
	if !strings.HasPrefix(url, srv.URL) {
		t.Errorf("url %q does not point at the test server", url)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := client.Complete(ctx, model.Request{
		Model:     "qwen3:14b",
		System:    []model.Block{{Text: strings.Repeat("primer ", 2000)}},
		User:      "0: Pass\n1: Cast Lightning Bolt\n",
		MaxTokens: 128,
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	// The probe must send exactly what a seat sends — it goes through
	// the same OpenAIClient — or it is diagnosing a request nobody
	// makes. Both thinking-off fields (#839) travel with it.
	if think, ok := got["think"].(bool); !ok || think {
		t.Errorf("request think field = %v (ok=%v), want false — the probe must send what the funnel sends", got["think"], ok)
	}
	if got["reasoning_effort"] != "none" {
		t.Errorf("request reasoning_effort = %v, want %q — the probe must send what the funnel sends", got["reasoning_effort"], "none")
	}

	res := probeResult{
		SystemBytes: len("primer ") * 2000,
		UserBytes:   len("0: Pass\n1: Cast Lightning Bolt\n"),
		Moves:       2,
	}
	res.PromptTokens = resp.Usage.InputTokens
	res.CompletionTokens = resp.Usage.OutputTokens
	res.CachedPromptTokens = resp.Usage.CachedPromptTokens
	res.FinishReason = resp.StopReason
	res.Reply = resp.Text
	res.Reasoning = resp.Reasoning
	res.ParsedIndex, res.ParseErr = model.ParseAnswerIndex(resp.Text)

	if res.PromptTokens != 6231 || res.CompletionTokens != 128 {
		t.Errorf("usage not decoded: %+v", res)
	}
	if res.CachedPromptTokens != 4096 {
		t.Errorf("prompt_tokens_details.cached_tokens = %d, want 4096", res.CachedPromptTokens)
	}
	if res.FinishReason != "length" {
		t.Errorf("finish_reason %q", res.FinishReason)
	}
	if res.Reasoning == "" {
		t.Fatal("the reasoning field did not survive decoding — this is the whole hypothesis")
	}
	if res.Reply != "" || res.ReplyParsed() {
		t.Errorf("content was empty on the wire; the probe thinks it parsed (%q, %v)", res.Reply, res.ParsedIndex)
	}
	if v := thinkingVerdict(res); !strings.HasPrefix(v, "THINKING: not suppressed") {
		t.Errorf("verdict on the measured qwen3 shape: %q", v)
	}
}

func TestBuildClientNeedsAnEndpoint(t *testing.T) {
	t.Setenv(model.EnvOpenAIEndpoint, "")
	if c, _ := buildClient(""); c != nil {
		t.Error("a client was built with no endpoint configured")
	}
	if c, url := buildClient("http://192.0.2.1:11434"); c == nil || !strings.HasSuffix(url, "/v1/chat/completions") {
		t.Errorf("the flag did not complete a bare host: %q", url)
	}
}

// representativeInput plays a real game to find a window worth
// probing with. It is the one part of the probe that can hang or come
// back empty, and it never runs in CI unless something asks it to —
// so something does.
func TestRepresentativeInputFindsAnEscalatedWindow(t *testing.T) {
	in, err := representativeInput(nil, "izzet-aggro")
	if err != nil {
		t.Fatalf("representativeInput: %v", err)
	}
	if len(in.Moves) < 5 {
		t.Errorf("window has %d moves; the probe wants one with a real choice in it", len(in.Moves))
	}
	if in.View.Turn.Number < 4 {
		t.Errorf("window is at turn %d; the probe wants a developed board", in.View.Turn.Number)
	}
	if in.Seat == uuid.Nil || in.View.ID == "" {
		t.Errorf("window has no seat or no view: %+v", in.View.Turn)
	}
	// And the prompt built from it is the funnel's, with the move list
	// in it — which is what makes the byte count mean anything.
	cfg := model.DefaultConfig()
	cfg.Routine.ID = "probe-model"
	req, _, v := model.New(cfg).BuildRequest(context.Background(), in)
	if v.Absorbed() {
		t.Fatal("the probe picked a window Layer A settles; the model would never see it")
	}
	if !strings.Contains(req.User, "0: ") {
		t.Errorf("the user delta has no numbered move list in it:\n%s", head(req.User, 400))
	}
	t.Logf("probe window: turn %d %s, %d moves, system %d bytes, user %d bytes (~%d tokens)",
		in.View.Turn.Number, in.View.Turn.Step, len(in.Moves),
		systemBytes(req), len(req.User), (systemBytes(req)+len(req.User))/estimateDivisor)
}
