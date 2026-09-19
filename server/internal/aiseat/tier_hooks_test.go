package aiseat_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// tier_hooks_test.go is #1060 seen from the table rather than from a
// type assertion.
//
// tiers/capability_test.go asks whether a factory-built policy IS a
// TargetOrderer and a CostFuelPricer. That is the assertion whose
// absence was the bug, but on its own it is still a claim about
// interfaces. This file asks the question a player would: does a seat
// the LOBBY built get a different, better move list because of them?
//
// Both tests run a real aiseat.Runner over a real room with a real
// factory-built policy, and compare the move list the runner handed
// the policy against the SAME board enumerated with no hooks at all.
// Before #1060 the two were identical on every shipped tier.

// --- the watcher -----------------------------------------------------

// orderWatcher wraps a factory-built tier policy, records the move
// list the runner enumerated for it, and — for the same board, at the
// same instant — the move list the enumerator produces with no
// ordering hooks. The pair is the evidence.
//
// It declares Unwrap, which is the whole of what a wrapper has to do
// for the hooks underneath it to keep working (aiseat/capability.go).
// A wrapper here that did NOT would make this test pass vacuously by
// reproducing the very defect it is checking for, so the Unwrap is
// load-bearing rather than decoration.
//
// Holding a *game.Game is a test-double liberty, not a policy one:
// ADR 0033 §3's ban is on the policy PACKAGES under aiseat/, and the
// baseline enumeration is the thing being compared against rather
// than anything the seat plays on. The wrapped policy still sees only
// the filtered Input the runner built.
type orderWatcher struct {
	inner aiseat.Policy
	g     *game.Game
	seat  uuid.UUID
	// want selects the window to capture: the first one it returns
	// true for is the one recorded.
	want func(moves []legal.Move) bool

	mu       sync.Mutex
	captured bool
	ordered  []legal.Move
	baseline []legal.Move
}

func (w *orderWatcher) Name() string { return w.inner.Name() }

// Unwrap is what keeps the wrapped tier's optional hooks reachable.
func (w *orderWatcher) Unwrap() aiseat.Policy { return w.inner }

func (w *orderWatcher) Decide(ctx context.Context, in aiseat.Input) (aiseat.Decision, error) {
	w.mu.Lock()
	if !w.captured && w.want(in.Moves) {
		w.ordered = append([]legal.Move(nil), in.Moves...)
		// The same board, enumerated the way every non-bot caller
		// does: no OrderTargets, no OrderCostFuel. Taken here, inside
		// the decision, because this is the one moment the board is
		// guaranteed to be the board that produced in.Moves.
		w.baseline = legal.EnumerateFor(w.g, w.seat)
		w.captured = true
	}
	w.mu.Unlock()
	return w.inner.Decide(ctx, in)
}

func (w *orderWatcher) snapshot() (ordered, baseline []legal.Move, ok bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.ordered, w.baseline, w.captured
}

// targetsOf reads a cast move's target ids off the wire payload, the
// way a policy would.
func targetsOf(m legal.Move) []string {
	var p struct {
		Targets []struct {
			ID string `json:"id"`
		} `json:"targets"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		return nil
	}
	out := make([]string, 0, len(p.Targets))
	for _, t := range p.Targets {
		out = append(out, t.ID)
	}
	return out
}

// rankOf is the position of the first move that targets id, or -1.
func rankOf(moves []legal.Move, source uuid.UUID, id uuid.UUID) int {
	n := 0
	for _, m := range moves {
		if m.Kind != legal.KindCast || m.Source != source {
			continue
		}
		for _, got := range targetsOf(m) {
			if got == id.String() {
				return n
			}
		}
		n++
	}
	return -1
}

// seedHand puts a card in a seat's hand, known to its owner alone,
// the way a drawn card is.
func seedHand(g *game.Game, owner *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = owner.ID, owner.ID
	c.KnownBy = map[uuid.UUID]bool{owner.ID: true}
	owner.Hand.PushTop(c)
	return c.InstanceID
}

// --- #687: the seat's target order reaches the enumerator ------------

// A lobby-built `heuristic` seat points its removal at the board's
// biggest threat FIRST, because its TargetOrder reached
// legal.Options.OrderTargets.
//
// The fixture puts the small creature into the battlefield first, so
// the engine's own candidate order offers it first. That is the order
// every seat got before #1060 — and with MaxExpansionPerSource at 12
// on a wide board it is the order in which the cap throws targets
// away, which is why #687 exists at all.
func TestALobbySeatOrdersItsTargetsByThreat(t *testing.T) {
	room := newRoom(t, 2, 10600)
	g := room.Game
	bot, opp := g.Seats[0], g.Seats[1]

	// The squire goes in first: zone order, and so candidate order,
	// puts the worthless creature ahead of the real threat.
	squire := seedBattlefield(g, opp, game.Card{
		Name: "Squire", TypeLine: "Creature — Human Soldier", Power: 1, Toughness: 1,
	})
	colossus := seedBattlefield(g, opp, game.Card{
		Name: "Colossus", TypeLine: "Creature — Golem", Power: 7, Toughness: 7,
	})
	seedBattlefield(g, bot, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	bolt := seedHand(g, bot, game.Card{
		Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleLightningBolt,
	})

	// The seat the LOBBY would build, not a bare heuristic.
	inner, err := tiers.NewFactory(tiers.FactoryOptions{}).NewPolicy(
		aiseat.SeatSpec{PlayerID: bot.ID, Tier: string(aiseat.TierHeuristic)})
	if err != nil {
		t.Fatalf("NewPolicy: %v", err)
	}
	watch := &orderWatcher{
		inner: inner, g: g, seat: bot.ID,
		want: func(moves []legal.Move) bool {
			return rankOf(moves, bolt, squire) >= 0 && rankOf(moves, bolt, colossus) >= 0
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := aiseat.Start(ctx, room, bot.ID, watch, aiseat.Config{}, nil, testLogger())
	h := aiseat.Start(ctx, room, opp.ID, &scripted{prefer: []string{"Keep hand"}}, aiseat.Config{}, nil, testLogger())

	waitFor(t, "a window offering the bolt at both creatures", func() bool {
		_, _, ok := watch.snapshot()
		return ok
	})
	cancel()
	waitForRunner(t, "the bot to stop", r)
	waitForRunner(t, "the opponent to stop", h)

	ordered, baseline, _ := watch.snapshot()
	gotBig, gotSmall := rankOf(ordered, bolt, colossus), rankOf(ordered, bolt, squire)
	if gotBig < 0 || gotSmall < 0 {
		t.Fatalf("the captured window offered the bolt at %d/%d of the two creatures", gotBig, gotSmall)
	}
	if gotBig > gotSmall {
		t.Errorf("the lobby seat was offered the 1/1 (rank %d) before the 7/7 (rank %d) — "+
			"#687's ordering did not reach the enumerator", gotSmall, gotBig)
	}

	// And the baseline: the SAME board with no hooks puts them the
	// other way round. Without this the test would still pass on a
	// board the engine happened to order correctly, and would then be
	// asserting nothing.
	baseBig, baseSmall := rankOf(baseline, bolt, colossus), rankOf(baseline, bolt, squire)
	if baseBig < 0 || baseSmall < 0 {
		t.Fatalf("the unordered enumeration offered the bolt at %d/%d of the two creatures", baseBig, baseSmall)
	}
	if baseSmall > baseBig {
		t.Fatalf("the fixture is broken: the engine's own candidate order already puts the 7/7 "+
			"(rank %d) ahead of the 1/1 (rank %d), so the ordered list proves nothing", baseBig, baseSmall)
	}
	t.Logf("ordered: 7/7 at %d, 1/1 at %d; unordered: 7/7 at %d, 1/1 at %d",
		gotBig, gotSmall, baseBig, baseSmall)
}

// --- #1013: the seat's fuel price reaches the enumerator -------------

// A lobby-built `assisted` seat escaping Uro is offered the payment
// that eats LANDS first, because its CostFuelPrice reached
// legal.Options.OrderCostFuel — through the model funnel, which is the
// wrapper #1060 found swallowing it.
//
// `assisted` rather than `heuristic` on purpose: the funnel is the
// second of the two wrappers, it has its own forwarding to get wrong,
// and with no transport it is a complete Layer A + B policy, so the
// opinion being forwarded is unambiguously the heuristic's.
func TestALobbySeatPricesItsEscapeFuel(t *testing.T) {
	room := newGraveyardRoom(t, 10601)
	g := room.Game
	bot, opp := g.Seats[0], g.Seats[1]

	// The spells go in FIRST, so zone order names them first and an
	// unpriced enumeration exiles three of them.
	spells := map[string]bool{}
	for _, name := range []string{"Ancestral Vision", "Snapcaster Bait", "Second Thoughts"} {
		id := seedGraveyard(g, bot, game.Card{Name: name, TypeLine: "Instant", ManaCost: "{1}{U}"})
		spells[id.String()] = true
	}
	for i := 0; i < 5; i++ {
		seedGraveyard(g, bot, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	}
	uro := seedGraveyard(g, bot, game.Card{
		Name: "Uro, Titan of Nature's Wrath", TypeLine: "Legendary Creature — Elder Giant",
		Power: 6, Toughness: 6, ManaCost: "{1}{G}{U}", OracleID: oracleUroBot,
	})
	for i := 0; i < 3; i++ {
		seedBattlefield(g, bot, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
		seedBattlefield(g, bot, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	}

	inner, err := tiers.New(tiers.Assisted, tiers.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, isFunnel := inner.(*model.Policy); !isFunnel {
		t.Fatalf("the assisted tier is a %T; this test is about the funnel's forwarding", inner)
	}
	watch := &orderWatcher{
		inner: inner, g: g, seat: bot.ID,
		want: func(moves []legal.Move) bool { return firstEscapePayment(moves, uro) != nil },
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := aiseat.Start(ctx, room, bot.ID, watch, aiseat.Config{MaxThink: 2 * time.Second}, nil, testLogger())
	h := aiseat.Start(ctx, room, opp.ID, passPolicy{}, aiseat.Config{}, nil, testLogger())

	waitFor(t, "a window offering Uro's escape", func() bool {
		_, _, ok := watch.snapshot()
		return ok
	})
	cancel()
	waitForRunner(t, "the bot to stop", r)
	waitForRunner(t, "the opponent to stop", h)

	ordered, baseline, _ := watch.snapshot()
	payment := firstEscapePayment(ordered, uro)
	if len(payment) != 5 {
		t.Fatalf("the first escape offered names %d cards, the cost is five: %v", len(payment), payment)
	}
	for _, id := range payment {
		if spells[id] {
			t.Errorf("the first escape payment offered to a lobby seat eats a SPELL (%s) with five "+
				"Forests in the same graveyard — #1013's price did not reach the enumerator: %v", id, payment)
		}
	}

	// The baseline: unpriced, the same board offers a payment that
	// eats the spells, because zone order put them first.
	base := firstEscapePayment(baseline, uro)
	if len(base) == 0 {
		t.Fatalf("the unordered enumeration offered no escape at all")
	}
	ateASpell := false
	for _, id := range base {
		if spells[id] {
			ateASpell = true
		}
	}
	if !ateASpell {
		t.Fatalf("the fixture is broken: the unpriced enumeration already spares the spells (%v), "+
			"so the priced one proves nothing", base)
	}
	t.Logf("priced payment %v; unpriced payment %v", payment, base)
}

// firstEscapePayment is the alternative-cost payment of the first
// offered cast of src, or nil when none is offered.
func firstEscapePayment(moves []legal.Move, src uuid.UUID) []string {
	for _, m := range moves {
		if m.Kind != legal.KindCast || m.Source != src {
			continue
		}
		if ids := altCostIDsOfCast(m); len(ids) > 0 {
			return ids
		}
	}
	return nil
}
