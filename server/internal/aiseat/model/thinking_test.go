package model

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// thinking_test.go pins the thinking experiment (#2196): an opt-in
// mode that lets a local model think before it picks a move. The
// default must not move, the opt-in must actually stop sending the
// thinking-off fields and raise the reply budget, and the answer is
// read from `content` and never from the thinking.

// captureOpenAI serves one canned reply and keeps every request body.
func captureOpenAI(t *testing.T, reply string) (*OpenAIClient, *[]map[string]any) {
	t.Helper()
	var bodies []map[string]any
	c := serveOpenAI(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var got map[string]any
		_ = json.Unmarshal(raw, &got)
		bodies = append(bodies, got)
		_, _ = w.Write([]byte(reply))
	})
	return c, &bodies
}

const answerBear = `{"choices":[{"message":{"content":"{\"index\": 2, \"move\": \"Cast Bear\", \"why\": \"curve out\"}"},"finish_reason":"stop"}]}`

func TestThinkingModeSendsNoThinkingOffFieldsAndALargerBudget(t *testing.T) {
	c, bodies := captureOpenAI(t, answerBear)
	p := testPolicy(t, c, &stubB{index: 0}, func(cfg *Config) {
		*cfg = cfg.WithThinking(0)
	})
	if d := decide(t, p, castWindow(), 2*time.Second); d.Index != 2 {
		t.Fatalf("index = %d, want the model's 2", d.Index)
	}
	if len(*bodies) != 1 {
		t.Fatalf("%d requests, want 1", len(*bodies))
	}
	got := (*bodies)[0]
	if v, ok := got["reasoning_effort"]; ok {
		t.Errorf("reasoning_effort = %v was sent; with thinking on it must be absent, or Ollama turns thinking off", v)
	}
	if got["think"] != true {
		t.Errorf("think = %v, want true", got["think"])
	}
	if got["max_tokens"] != float64(DefaultThinkingMaxTokens) {
		t.Errorf("max_tokens = %v, want %d — a thinking model writes its reasoning out of this budget", got["max_tokens"], DefaultThinkingMaxTokens)
	}
	// The #2196 constraint stays on: thinking was measured to work with
	// the schema on Ollama, and the schema is what keeps the answer on
	// the list.
	if _, ok := got["response_format"]; !ok {
		t.Error("response_format was dropped in thinking mode")
	}
}

func TestDefaultModeStillSuppressesThinkingWithTheSmallBudget(t *testing.T) {
	c, bodies := captureOpenAI(t, answerBear)
	p := testPolicy(t, c, &stubB{index: 0}, nil)
	decide(t, p, castWindow(), 2*time.Second)
	got := (*bodies)[0]
	if got["think"] != false || got["reasoning_effort"] != "none" {
		t.Errorf("think = %v, reasoning_effort = %v; the default must keep thinking OFF", got["think"], got["reasoning_effort"])
	}
	// castWindow escalates (two castable spells), so the frontier
	// profile answered it: 256, the shipped number.
	if mt := got["max_tokens"]; mt != float64(DefaultConfig().Frontier.MaxTokens) && mt != float64(DefaultConfig().Routine.MaxTokens) {
		t.Errorf("max_tokens = %v, want the shipped 128 or 256", mt)
	}
}

// Thinking that runs the budget out leaves content empty. That is the
// failure ADR 0052 §7 names, and it stays a malformed reply: nothing
// is read out of the reasoning.
func TestAReasoningOnlyReplyIsMalformed(t *testing.T) {
	c, _ := captureOpenAI(t, `{"choices":[{"message":{"content":"","reasoning":"I think the answer is {\"index\": 2} but"},"finish_reason":"length"}]}`)
	p := testPolicy(t, c, &stubB{index: 1, reason: "the heuristic's pick"}, func(cfg *Config) {
		*cfg = cfg.WithThinking(4000)
	})
	d, tr := decideTraced(t, p, castWindow())
	if d.Index != 1 {
		t.Fatalf("index = %d, want Layer B's 1 — an index in the reasoning is a draft, not an answer", d.Index)
	}
	if tr.Fallback != FallbackMalformed {
		t.Errorf("fallback = %q, want %q", tr.Fallback, FallbackMalformed)
	}
	if tr.ReasoningChars == 0 {
		t.Error("the reasoning's length was not recorded, so the run cannot say the budget went to thinking")
	}
}

func TestAReplyWithReasoningAndContentParses(t *testing.T) {
	const reasoning = "The Bear develops the board; the Bolt can wait."
	c, _ := captureOpenAI(t, `{"choices":[{"message":{"content":"{\"index\": 2, \"move\": \"Cast Bear\", \"why\": \"curve\"}","reasoning":"`+reasoning+`"},"finish_reason":"stop"}]}`)
	p := testPolicy(t, c, &stubB{index: 0}, func(cfg *Config) {
		*cfg = cfg.WithThinking(0)
	})
	d, tr := decideTraced(t, p, castWindow())
	if d.Index != 2 || tr.Layer != LayerC {
		t.Fatalf("index %d layer %q, want the model's 2 on Layer C", d.Index, tr.Layer)
	}
	if tr.ReasoningChars != len(reasoning) {
		t.Errorf("trace ReasoningChars = %d, want %d", tr.ReasoningChars, len(reasoning))
	}
	st := p.Stats()
	if st.ReasoningReplies != 1 || st.ReasoningChars != int64(len(reasoning)) {
		t.Errorf("stats reasoning = %d replies / %d chars, want 1 / %d", st.ReasoningReplies, st.ReasoningChars, len(reasoning))
	}
	if recs := p.Records(); len(recs) != 1 || recs[0].ReasoningChars != len(reasoning) {
		t.Errorf("record = %+v", recs)
	}
}

// A server that sends the thinking inline, ahead of the answer, must
// not have a brace in the draft read as the answer.
func TestAnInlineThinkBlockIsNotTheAnswer(t *testing.T) {
	for _, tc := range []struct {
		name, content, wantText, wantReasoning string
	}{
		{
			name:          "closed",
			content:       `<think>maybe {\"index\": 1}</think>\n{\"index\": 2}`,
			wantText:      `{"index": 2}`,
			wantReasoning: `maybe {"index": 1}`,
		},
		{
			name:          "unclosed — the budget ran out mid-thought",
			content:       `<think>still going {\"index\": 1}`,
			wantText:      ``,
			wantReasoning: `still going {"index": 1}`,
		},
		{
			name:     "no block",
			content:  `{\"index\": 2}`,
			wantText: `{"index": 2}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := captureOpenAI(t, `{"choices":[{"message":{"content":"`+tc.content+`"},"finish_reason":"stop"}]}`)
			resp, err := c.Complete(context.Background(), Request{Model: "m", User: "u"})
			if err != nil {
				t.Fatal(err)
			}
			if resp.Text != tc.wantText || resp.Reasoning != tc.wantReasoning {
				t.Errorf("text %q reasoning %q, want %q / %q", resp.Text, resp.Reasoning, tc.wantText, tc.wantReasoning)
			}
		})
	}
}

func TestWithThinkingTouchesOnlyTheDecisionProfiles(t *testing.T) {
	base := DefaultConfig()
	on := base.WithThinking(0)
	if on.Routine.Thinking != ThinkingAdaptive || on.Frontier.Thinking != ThinkingAdaptive {
		t.Errorf("thinking = %q / %q, want adaptive on both decision profiles", on.Routine.Thinking, on.Frontier.Thinking)
	}
	if on.Routine.MaxTokens != DefaultThinkingMaxTokens || on.Frontier.MaxTokens != DefaultThinkingMaxTokens {
		t.Errorf("max tokens = %d / %d, want %d", on.Routine.MaxTokens, on.Frontier.MaxTokens, DefaultThinkingMaxTokens)
	}
	if on.Improv != base.Improv {
		t.Errorf("Improv changed: %+v, want %+v", on.Improv, base.Improv)
	}
	if got := base.WithThinking(4000); got.Routine.MaxTokens != 4000 || got.Frontier.MaxTokens != 4000 {
		t.Errorf("an explicit budget was not applied: %+v / %+v", got.Routine, got.Frontier)
	}
	// The value receiver is the point: the shipped default is never
	// edited in place.
	if base.Routine.Thinking != "" || base.Routine.MaxTokens != 128 {
		t.Errorf("DefaultConfig was mutated: %+v", base.Routine)
	}
	if got := base.WithMaxTokens(512); got.Routine.MaxTokens != 512 || got.Routine.Thinking != "" {
		t.Errorf("WithMaxTokens = %+v, want 512 and thinking untouched", got.Routine)
	}
	if got := base.WithMaxTokens(0); got.Routine != base.Routine || got.Frontier != base.Frontier {
		t.Error("WithMaxTokens(0) changed something")
	}
}

func TestThinkingFromEnv(t *testing.T) {
	for _, tc := range []struct {
		think, tokens string
		wantThink     bool
		wantTokens    int
		wantErr       bool
	}{
		{},
		{think: "0"},
		{think: "1", wantThink: true},
		{think: "on", tokens: "4000", wantThink: true, wantTokens: 4000},
		{tokens: "512", wantTokens: 512},
		{think: "maybe", wantErr: true},
		{tokens: "lots", wantErr: true},
		{tokens: "-1", wantErr: true},
	} {
		t.Setenv(EnvBotThink, tc.think)
		t.Setenv(EnvBotMaxTokens, tc.tokens)
		think, tokens, err := ThinkingFromEnv()
		if (err != nil) != tc.wantErr {
			t.Errorf("think=%q tokens=%q: err = %v, wantErr %v", tc.think, tc.tokens, err, tc.wantErr)
			continue
		}
		if err == nil && (think != tc.wantThink || tokens != tc.wantTokens) {
			t.Errorf("think=%q tokens=%q: got %v / %d, want %v / %d", tc.think, tc.tokens, think, tokens, tc.wantThink, tc.wantTokens)
		}
		if err != nil && !strings.Contains(err.Error(), "CMDCTRL_BOT_") {
			t.Errorf("error %q does not name the variable", err)
		}
	}
}
