package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/suite"
)

type deadlineCaptureTransport func(*http.Request) (*http.Response, error)

func (f deadlineCaptureTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestSuiteDefaultDeadlineReachesModelTransport(t *testing.T) {
	t.Setenv("CMDCTRL_BOT_MAX_THINK", "")
	t.Setenv("CMDCTRL_OPENAI_ENDPOINT", "http://review.invalid")
	t.Setenv("CMDCTRL_BOT_MODEL", "review-model")
	var budget time.Duration
	previous := http.DefaultTransport
	http.DefaultTransport = deadlineCaptureTransport(func(r *http.Request) (*http.Response, error) {
		deadline, _ := r.Context().Deadline()
		budget = time.Until(deadline)
		return &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"choices":[{"message":{"content":"0"},"finish_reason":"stop"}]}`)),
		}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = previous })

	opts, err := parseSuiteRun([]string{"--policy", "assisted"})
	if err != nil {
		t.Fatal(err)
	}
	policy, err := buildSuitePolicy(os.Stdout, opts)
	if err != nil {
		t.Fatal(err)
	}
	pos, err := suite.LoadFile("../../internal/aiseat/suite/testdata/positions/attack-an-empty-board.json")
	if err != nil {
		t.Fatal(err)
	}
	suite.Run(context.Background(), []suite.Position{pos}, policy, suite.RunOptions{MaxThink: opts.MaxThink})
	if budget < 50*time.Second {
		t.Fatalf("suite deadline is %v but transport only received %v", opts.MaxThink, budget)
	}
}

func TestSuiteNoBudgetDoesNotCountAsModelAgreement(t *testing.T) {
	t.Setenv("CMDCTRL_BOT_MAX_THINK", "")
	t.Setenv("CMDCTRL_OPENAI_ENDPOINT", "http://review.invalid")
	t.Setenv("CMDCTRL_BOT_MODEL", "review-model")
	opts, err := parseSuiteRun([]string{"--policy", "assisted", "--max-think", "1ms"})
	if err != nil {
		t.Fatal(err)
	}
	policy, err := buildSuitePolicy(os.Stdout, opts)
	if err != nil {
		t.Fatal(err)
	}
	pos, err := suite.LoadFile("../../internal/aiseat/suite/testdata/positions/attack-an-empty-board.json")
	if err != nil {
		t.Fatal(err)
	}
	rep := suite.Run(context.Background(), []suite.Position{pos}, policy, suite.RunOptions{MaxThink: opts.MaxThink})
	if rep.Agree != 0 || rep.Errors != 1 {
		t.Fatalf("no-budget fallback was graded as model agreement: %+v", rep.Results[0])
	}
}
