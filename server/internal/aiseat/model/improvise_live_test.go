package model

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// improvise_live_test.go is the one test in this package that talks to
// a network, and it is why every other one does not.
//
// A FakeClient proves the PLUMBING — the window, the prompt, the
// parse, the cap, the four failure paths — and proves nothing
// whatsoever about whether a real model, shown a real card's oracle
// text, writes a bundle that the rail will accept. That is the
// question this test asks, and it can only be asked with a key.
//
// CI never sets AISEAT_LIVE_MODEL_TESTS. Run it by hand:
//
//	AISEAT_LIVE_MODEL_TESTS=1 CMDCTRL_ANTHROPIC_API_KEY=sk-... \
//	  go test ./internal/aiseat/model/ -run TestLive -v
//
// A local endpoint works too: set CMDCTRL_OPENAI_ENDPOINT and
// CMDCTRL_BOT_MODEL instead. CMDCTRL_BOT_FRONTIER_MODEL, if set,
// chooses the model — improvisation asks the frontier slot.
const envLiveModelTests = "AISEAT_LIVE_MODEL_TESTS"

// liveClient builds the transport the deployment would build, or
// skips. The selection order is main.go's: a named local endpoint
// beats a key lying around in the environment.
func liveClient(t *testing.T) (Client, string) {
	t.Helper()
	if os.Getenv(envLiveModelTests) == "" {
		t.Skipf("live model test: set %s=1 (and a key) to run; CI never does", envLiveModelTests)
	}
	id := strings.TrimSpace(os.Getenv("CMDCTRL_BOT_FRONTIER_MODEL"))
	if id == "" {
		id = strings.TrimSpace(os.Getenv("CMDCTRL_BOT_MODEL"))
	}
	if oc := NewOpenAIClient(); oc != nil {
		if id == "" {
			t.Skip("CMDCTRL_OPENAI_ENDPOINT is set but CMDCTRL_BOT_MODEL is not; a local endpoint needs the served model's name")
		}
		return oc, id
	}
	if ac := NewAnthropicClient(); ac != nil {
		if id == "" {
			id = DefaultConfig().Improv.ID
		}
		return ac, id
	}
	t.Skipf("%s is set but no transport is: set CMDCTRL_ANTHROPIC_API_KEY or CMDCTRL_OPENAI_ENDPOINT", envLiveModelTests)
	return nil, ""
}

// TestLiveModelImprovisesARealCard sends one real improvisation call
// and checks the answer against the rail the runner would put it
// through.
//
// It accepts a DECLINE as a pass. "Each opponent loses 2 life and you
// gain 2 life" is squarely inside the four verbs, so a decline would
// be a disappointing answer — but it is a CORRECT one, and a test
// that failed on it would be pressuring a live model to act when the
// whole design says declining is cheap and safe. What the test
// refuses to accept is a bundle that is wrong: a verb outside the
// four, an id nobody has, a shape the rail throws away.
func TestLiveModelImprovisesARealCard(t *testing.T) {
	client, id := liveClient(t)

	p := New(Config{
		Tier:      TierStrong,
		Fallback:  &stubB{index: 0, reason: "stub"},
		Client:    client,
		Improvise: true,
		Improv:    ModelProfile{ID: id, MaxTokens: 1024},
		Deck:      improvDeck(),
		Log:       testLogger(),
		MaxCall:   60 * time.Second,
	})

	// The same two windows the fake tests use: the spell crosses the
	// stack, then it is in the graveyard having done nothing.
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	p.Improvise(ctx, improvInput(castingView(hexID, "Hex Drain", "Instant")))
	resolved := resolvedInGraveyard(hexID, "Hex Drain", "Instant")
	im, ok := p.Improvise(ctx, improvInput(resolved))

	st := p.Stats()
	t.Logf("model=%s outcomes=%v usage=%+v latency=%v", id, st.ByImprov, st.ImprovUsage, st.ImprovLatency)
	if !ok {
		if st.ByImprov[ImprovDeclined] == 1 {
			t.Logf("the model declined; that is a correct answer, if a dull one")
			return
		}
		t.Fatalf("no bundle and no decline: %v", st.ByImprov)
	}

	if im.Card != "Hex Drain" {
		t.Errorf("Card = %q", im.Card)
	}
	if err := im.Validate(); err != nil {
		t.Fatalf("the live model's bundle does not pass the rail: %v\nsteps: %+v", err, im.Steps)
	}
	for i, s := range im.Steps {
		t.Logf("step %d: %s player=%s params=%s", i, s.Type, s.Player, s.Params)
	}
	if bad := unknownRefs(im, resolved); len(bad) > 0 {
		t.Errorf("the bundle names ids that are not in this seat's view: %v\n"+
			"the rail would roll it back, but the prompt is meant to make this impossible", bad)
	}
	t.Logf("announcement: %s", im.Announcement())
}

// unknownRefs returns every instance_id / player the bundle names
// that does not appear in the view. A bundle naming one would be
// rolled back by Room.ApplyBundle — this is checking the PROMPT, not
// the rail: the primer promises the model that only listed ids work,
// and a live model reaching past them means the prompt is not saying
// it clearly enough.
func unknownRefs(im aiseat.Improvisation, v protocol.GameView) []string {
	known := map[string]bool{}
	for i := range v.Battlefield.Cards {
		known[v.Battlefield.Cards[i].InstanceID] = true
	}
	for i := range v.Exile.Cards {
		known[v.Exile.Cards[i].InstanceID] = true
	}
	for i := range v.Stack.Cards {
		known[v.Stack.Cards[i].InstanceID] = true
	}
	for i := range v.Seats {
		s := &v.Seats[i]
		known[s.ID] = true
		for _, z := range [][]protocol.CardView{s.Hand.Cards, s.Graveyard.Cards, s.Command.Cards} {
			for j := range z {
				known[z[j].InstanceID] = true
			}
		}
	}

	var bad []string
	for _, s := range im.Steps {
		if s.Player != uuid.Nil && !known[s.Player.String()] {
			bad = append(bad, s.Player.String())
		}
		var p struct {
			InstanceID string `json:"instance_id"`
			Src        struct {
				Owner string `json:"owner"`
			} `json:"src"`
			Dst struct {
				Owner string `json:"owner"`
			} `json:"dst"`
		}
		if err := json.Unmarshal(s.Params, &p); err != nil {
			continue
		}
		for _, ref := range []string{p.InstanceID, p.Src.Owner, p.Dst.Owner} {
			if ref != "" && !known[ref] {
				bad = append(bad, ref)
			}
		}
	}
	return bad
}
