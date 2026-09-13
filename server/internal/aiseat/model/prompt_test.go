package model

import (
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func testDeck() DeckProfile {
	return DeckProfile{
		Name:      "Mono-red aggro (test)",
		Archetype: "Curve out, point burn at whoever is winning, attack every turn.",
		Cards: []DeckCard{
			{Name: "Lightning Bolt", Cost: "{R}", Type: "Instant", Oracle: "Lightning Bolt deals 3 damage to any target."},
			{Name: "Bear", Cost: "{1}{R}", Type: "Creature — Bear", Oracle: ""},
			{Name: "Mountain", Type: "Basic Land — Mountain", Oracle: "{T}: Add {R}."},
		},
	}
}

// --- the static half ------------------------------------------------

// The cache is a PREFIX match, so the static block has to be
// byte-identical on every call of the game. This is the test that
// catches a timestamp or a turn number slipping into it — a bug with
// no symptom other than a bill.
func TestStaticBlockIsByteIdenticalAcrossDecisions(t *testing.T) {
	fake := AlwaysIndex(0)
	p := testPolicy(t, fake, &stubB{index: 0}, func(c *Config) { c.Deck = testDeck() })

	for i := 0; i < 4; i++ {
		in := castWindow()
		in.View.Turn.Number = 3 + i // the board moves; the prefix must not
		in.View.Seats[1].Life = 31 - i
		decide(t, p, in, 2*time.Second)
	}
	reqs := fake.Requests()
	if len(reqs) != 4 {
		t.Fatalf("calls = %d, want 4", len(reqs))
	}
	first := reqs[0].System
	for i, r := range reqs[1:] {
		if len(r.System) != len(first) {
			t.Fatalf("call %d has %d system blocks, want %d", i+1, len(r.System), len(first))
		}
		for j := range first {
			if r.System[j].Text != first[j].Text {
				t.Errorf("call %d system block %d differs from the first — the prompt cache is dead", i+1, j)
			}
		}
	}
	// And the deltas must NOT be identical, or the test above is
	// measuring a prompt that carries no board at all.
	if reqs[0].User == reqs[1].User {
		t.Error("two different boards produced the same delta; the prompt is not carrying the board")
	}
}

func TestStaticBlockCarriesExactlyOneCacheBreakpointAtTheEnd(t *testing.T) {
	blocks := testDeck().staticBlocks()
	if len(blocks) < 2 {
		t.Fatalf("blocks = %d, want the primer and the decklist", len(blocks))
	}
	marks := 0
	for i, b := range blocks {
		if b.Cache {
			marks++
			if i != len(blocks)-1 {
				t.Errorf("the breakpoint is on block %d of %d; it belongs on the last one or the rest is uncached",
					i, len(blocks))
			}
		}
	}
	if marks != 1 {
		t.Errorf("%d cache breakpoints, want 1", marks)
	}
}

// The decklist is normalised, so a caller that assembles it in a
// different order does not silently pay for a cache miss.
func TestDecklistOrderDoesNotChangeTheStaticBlock(t *testing.T) {
	a := testDeck()
	b := testDeck()
	b.Cards[0], b.Cards[2] = b.Cards[2], b.Cards[0]
	ab, bb := a.staticBlocks(), b.staticBlocks()
	if len(ab) != len(bb) {
		t.Fatalf("block counts differ: %d vs %d", len(ab), len(bb))
	}
	for i := range ab {
		if ab[i].Text != bb[i].Text {
			t.Errorf("block %d differs after reordering the decklist", i)
		}
	}
}

// An empty DeckProfile still has to produce a usable prompt: a bot
// seat with no configured deck profile is a misconfiguration, not a
// crash.
func TestEmptyDeckProfileStillProducesAPrimer(t *testing.T) {
	blocks := DeckProfile{}.staticBlocks()
	if len(blocks) != 1 || !strings.Contains(blocks[0].Text, "ONE NUMBER") {
		t.Fatalf("blocks = %+v", blocks)
	}
	if !blocks[0].Cache {
		t.Error("the only block is the prefix and it is not cached")
	}
}

// --- the per-decision delta -------------------------------------------

func TestDeltaCarriesTheMoveListWithRealIndices(t *testing.T) {
	fake := AlwaysIndex(0)
	p := testPolicy(t, fake, &stubB{index: 2, reason: "b"}, nil)
	decide(t, p, castWindow(), 2*time.Second)

	user := fake.Requests()[0].User
	for i, want := range []string{"0: Pass priority", "1: Cast Lightning Bolt", "2: Cast Bear"} {
		if !strings.Contains(user, want) {
			t.Errorf("delta is missing move %d (%q):\n%s", i, want, user)
		}
	}
	if !strings.Contains(user, "rule-based fallback") {
		t.Errorf("the delta does not tell the model what the fallback would do:\n%s", user)
	}
	if !strings.Contains(user, "TURN 7") || !strings.Contains(user, "40 life") {
		t.Errorf("the delta is missing the board:\n%s", user)
	}
}

// A prompt is a thing a person has to be able to read before they can
// judge it, and there is nowhere else in the codebase to look at one.
// Run with -v to see what a model is actually being sent.
func TestPromptIsLegible(t *testing.T) {
	fake := AlwaysIndex(0)
	p := testPolicy(t, fake, &stubB{index: 1, reason: "b"}, func(c *Config) { c.Deck = testDeck() })
	in := castWindow()
	in.View.Battlefield.Cards = []protocol.CardView{
		{InstanceID: "l1", Name: "Mountain", Controller: meSeat.String(), TypeLine: "Basic Land — Mountain", Tapped: true},
		{InstanceID: "c1", Name: "Bear", Controller: meSeat.String(), TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
			Counters: map[string]int{"+1/+1": 1}},
		{InstanceID: "c2", Name: "Wurm", Controller: oppSeat.String(), TypeLine: "Creature — Wurm", Power: 7, Toughness: 7,
			Abilities: []string{"trample"}, AttackingTarget: meSeat.String()},
	}
	decide(t, p, in, 2*time.Second)
	reqs := fake.Requests()
	for i, b := range reqs[0].System {
		t.Logf("--- system block %d (cached=%v) ---\n%s", i, b.Cache, b.Text)
	}
	t.Logf("--- per-decision delta ---\n%s", reqs[0].User)
}

// The hidden-information guarantee, at the one place it could leak
// without anyone noticing: the prompt. A policy cannot SEE an
// opponent's hand — the view is already filtered — but a renderer
// that walked every seat's Hand.Cards would still print whatever
// happened to be there, and a future change to the filter would turn
// that into a live leak. So the renderer prints card names for the
// bot's own zones only, and this is the test that says so.
func TestDeltaPrintsNoOpponentHandCards(t *testing.T) {
	fake := AlwaysIndex(0)
	p := testPolicy(t, fake, &stubB{index: 0}, nil)
	in := castWindow()
	// Plant a card in the opponent's hand that the real filter would
	// never have put there.
	in.View.Seats[1].Hand.Cards = []protocol.CardView{
		{InstanceID: "secret", Name: "Cyclonic Rift", ManaCost: "{1}{U}", TypeLine: "Instant"},
	}
	decide(t, p, in, 2*time.Second)

	user := fake.Requests()[0].User
	if strings.Contains(user, "Cyclonic Rift") {
		t.Errorf("the prompt named a card in an opponent's hand:\n%s", user)
	}
	if !strings.Contains(user, "4 cards in hand") {
		t.Errorf("the prompt should still carry the opponent's hand COUNT:\n%s", user)
	}
}

// A face-down card the seat is not a knower of has no name on the
// wire; the prompt must say so rather than render an empty string the
// model will fill in with an assumption.
func TestDeltaNamesRedactedCardsHonestly(t *testing.T) {
	fake := AlwaysIndex(0)
	p := testPolicy(t, fake, &stubB{index: 0}, nil)
	in := castWindow()
	in.View.Battlefield.Cards = []protocol.CardView{
		{InstanceID: "fd", Controller: oppSeat.String(), FaceDown: true, TypeLine: "Creature", Power: 2, Toughness: 2},
	}
	decide(t, p, in, 2*time.Second)
	if user := fake.Requests()[0].User; !strings.Contains(user, "a face-down card") {
		t.Errorf("a redacted permanent was not labelled:\n%s", user)
	}
}

// The cap has to keep the pass and the fallback's own pick, or the
// model loses the ability to do nothing and the prompt stops telling
// it what it is overruling.
func TestTheCappedMoveListAlwaysKeepsThePassAndTheFallback(t *testing.T) {
	moves := []legal.Move{pass()}
	for i := 0; i < 40; i++ {
		moves = append(moves, mv(legal.KindCast, "Cast filler", `{"instance_id":"x"}`))
	}
	in := aiseat.Input{Seat: meSeat, View: view(), Moves: moves}

	p := New(Config{Fallback: &stubB{index: 37}, Log: testLogger(), MaxCandidates: 5})
	_, shown := p.buildDelta(in, nil, 37)
	if len(shown) != 5 {
		t.Fatalf("shown = %d, want the cap of 5", len(shown))
	}
	var hasPass, hasFallback bool
	for _, s := range shown {
		if s.index == 0 {
			hasPass = true
		}
		if s.index == 37 {
			hasFallback = true
		}
	}
	if !hasPass {
		t.Error("the pass was capped out; the model can no longer choose to do nothing")
	}
	if !hasFallback {
		t.Error("the fallback's own move was capped out")
	}
}

// The shown list is ordered by index, so the prompt reads as a list
// rather than as the scorer's recommendation.
func TestShownMovesAreOrderedByIndex(t *testing.T) {
	p := New(Config{Fallback: &stubB{}, Log: testLogger()})
	in := castWindow()
	_, shown := p.buildDelta(in, nil, 0)
	for i := 1; i < len(shown); i++ {
		if shown[i-1].index >= shown[i].index {
			t.Fatalf("shown list is not ascending: %+v", shown)
		}
	}
}

// A deterministic prompt matters for the same reason the static block
// does, and Go randomises map iteration — counters are the one map
// that reaches the renderer.
func TestDeltaIsDeterministicForOneBoard(t *testing.T) {
	p := New(Config{Fallback: &stubB{}, Log: testLogger()})
	in := castWindow()
	in.View.Battlefield.Cards = []protocol.CardView{{
		InstanceID: "c", Name: "Bear", Controller: meSeat.String(), TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2,
		Counters: map[string]int{"+1/+1": 2, "stun": 1, "lore": 3, "charge": 4, "oil": 5},
	}}
	first, _ := p.buildDelta(in, nil, 0)
	for i := 0; i < 24; i++ {
		got, _ := p.buildDelta(in, nil, 0)
		if got != first {
			t.Fatalf("two renders of one board differ:\n%s\n---\n%s", first, got)
		}
	}
}

func TestEchoesFallbackReadsTheRenderedMoveList(t *testing.T) {
	fake := EchoesFallback()
	p := testPolicy(t, fake, &stubB{index: 2, reason: "b"}, nil)
	d := decide(t, p, castWindow(), 2*time.Second)
	if d.Index != 2 {
		t.Fatalf("index = %d, want the fallback's 2 echoed back through the prompt", d.Index)
	}
	if st := p.Stats(); st.ByLayer[LayerC] != 1 {
		t.Errorf("the echo did not land on Layer C: %+v", st.ByLayer)
	}
}

// A prompt the fake cannot read is a prompt a model would struggle
// with too; this pins the move-list format rather than letting it
// drift silently.
func TestIndexFromPromptFindsTheMarker(t *testing.T) {
	user := "noise\n  0: Pass priority\n  3: Cast Bear   <- the rule-based fallback would take this\n  4: Cast Ogre\n"
	got, ok := indexFromPrompt(user)
	if !ok || got != 3 {
		t.Fatalf("indexFromPrompt = %d, %v; want 3, true", got, ok)
	}
	first, ok := indexFromPrompt("  7: Pass priority\n  9: Cast Bear\n")
	if !ok || first != 7 {
		t.Fatalf("indexFromPrompt without a marker = %d, %v; want the first listed, true", first, ok)
	}
	if _, ok := indexFromPrompt("no moves here"); ok {
		t.Error("indexFromPrompt invented a move list")
	}
}
