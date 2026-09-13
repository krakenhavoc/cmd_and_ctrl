package model

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// fake.go is a Client that does not need a network, and it is a
// production file rather than a _test.go one on purpose: the failure
// modes it reproduces — outage, timeout, malformed reply, an index
// that is not a move — are exercised from aiseat's own game harness
// one package up, and the S31 exit criteria include a **model-outage
// drill** that has to run in CI on a machine with no API key.
//
// It is also how Layer C is exercised at all in an environment with
// no model endpoint. That is worth stating plainly rather than
// burying: a FakeClient proves the funnel's PLUMBING — prompt
// assembly, model selection, deadline arithmetic, index validation,
// the fallback on every failure — and proves nothing whatsoever about
// whether a real model plays Magic well. The first is what this
// sub-PR can verify offline; the second needs a key and a game.

// FakeClient is a scriptable Client.
type FakeClient struct {
	// Reply produces the response for call n (0-based). Required.
	Reply func(n int, req Request) (Response, error)
	// Delay is slept before Reply is consulted, respecting ctx. Use
	// it to make a call overrun the policy's budget.
	Delay time.Duration
	// Label names the transport in logs.
	Label string

	mu       sync.Mutex
	calls    int
	requests []Request
}

// Name identifies the transport.
func (f *FakeClient) Name() string {
	if f.Label != "" {
		return f.Label
	}
	return "fake"
}

// Complete records the request and runs Reply.
func (f *FakeClient) Complete(ctx context.Context, req Request) (Response, error) {
	f.mu.Lock()
	n := f.calls
	f.calls++
	f.requests = append(f.requests, req)
	f.mu.Unlock()

	if f.Delay > 0 {
		t := time.NewTimer(f.Delay)
		defer t.Stop()
		select {
		case <-ctx.Done():
			return Response{}, ctx.Err()
		case <-t.C:
		}
	}
	if err := ctx.Err(); err != nil {
		return Response{}, err
	}
	if f.Reply == nil {
		return Response{}, ErrNoClient
	}
	return f.Reply(n, req)
}

// Calls is how many times Complete was entered.
func (f *FakeClient) Calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

// Requests returns every request the fake was given, in order. The
// prompt-cache tests read the System blocks off these.
func (f *FakeClient) Requests() []Request {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]Request(nil), f.requests...)
}

// --- ready-made doubles ---------------------------------------------

// AlwaysFails is the outage drill's client: every call fails, from
// the first one, for the whole game.
func AlwaysFails(err error) *FakeClient {
	if err == nil {
		err = ErrOutage
	}
	return &FakeClient{
		Label: "always-fails",
		Reply: func(int, Request) (Response, error) { return Response{}, err },
	}
}

// FailsAfter answers normally for n calls and then fails forever.
// This is the outage that matters operationally: the endpoint was
// there when the game started and went away in the middle of it.
//
// It borrows inner's Reply rather than wrapping the whole client, so
// inner's own Delay and call counters do not apply — the counting
// that matters is the returned client's.
func FailsAfter(n int, inner *FakeClient, err error) *FakeClient {
	if err == nil {
		err = ErrOutage
	}
	return &FakeClient{
		Label: "fails-after",
		Reply: func(call int, req Request) (Response, error) {
			if call >= n || inner == nil || inner.Reply == nil {
				return Response{}, err
			}
			return inner.Reply(call, req)
		},
	}
}

// NeverAnswers blocks until the caller's deadline expires. It is the
// slow-model case: no error, no reply, just the budget running out.
func NeverAnswers() *FakeClient {
	return &FakeClient{Label: "never-answers", Delay: time.Hour}
}

// AlwaysText replies with a fixed body. Feed it nonsense to exercise
// the malformed path.
func AlwaysText(text string) *FakeClient {
	return &FakeClient{
		Label: "always-text",
		Reply: func(int, Request) (Response, error) {
			return Response{Text: text, Usage: Usage{InputTokens: 1, OutputTokens: 1}}, nil
		},
	}
}

// AlwaysIndex replies with a well-formed answer naming one index.
// Out-of-range values are how the out-of-range path is tested.
func AlwaysIndex(i int) *FakeClient {
	return &FakeClient{
		Label: "always-index",
		Reply: func(int, Request) (Response, error) {
			return Response{
				Text:  fmt.Sprintf(`{"index": %d, "why": "fake"}`, i),
				Usage: Usage{InputTokens: 1, OutputTokens: 1},
			}, nil
		},
	}
}

// EchoesFallback reads the move list out of the prompt and answers
// with whichever entry is marked as the rule-based fallback's pick,
// or the first listed move when the marker is absent.
//
// It is the recorded-model stand-in a whole game can be played
// against: every answer is in range and legal, so the Layer C path —
// assemble, call, parse, validate, dispatch — runs end to end with no
// endpoint, and the game that comes out is a real game. What it does
// NOT do is play differently from the heuristic, so it measures the
// plumbing and never the model.
func EchoesFallback() *FakeClient {
	return &FakeClient{
		Label: "echoes-fallback",
		Reply: func(_ int, req Request) (Response, error) {
			idx, ok := indexFromPrompt(req.User)
			if !ok {
				return Response{}, fmt.Errorf("fake: no move list in prompt")
			}
			return Response{
				Text:  fmt.Sprintf(`{"index": %d, "why": "echoes the fallback"}`, idx),
				Usage: Usage{InputTokens: len(req.User) / 4, OutputTokens: 12},
			}, nil
		},
	}
}

// indexFromPrompt parses the rendered move list. It deliberately
// reads the same text a model would, so a change to prompt.go that
// made the list unreadable breaks a test rather than quietly
// degrading play.
func indexFromPrompt(user string) (int, bool) {
	first, found := 0, false
	for _, line := range strings.Split(user, "\n") {
		t := strings.TrimSpace(line)
		colon := strings.IndexByte(t, ':')
		if colon <= 0 {
			continue
		}
		n, err := strconv.Atoi(t[:colon])
		if err != nil {
			continue
		}
		if !found {
			first, found = n, true
		}
		if strings.Contains(t, "rule-based fallback") {
			return n, true
		}
	}
	return first, found
}
