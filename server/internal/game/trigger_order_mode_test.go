package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// trigger_order_mode_test.go — a seat's trigger-order mode (#1530's
// "always ask", three modes since #1968) and #1968's widened skip:
// copies of one source-blind catalog ability need no CR 603.3b order.

// queueTwoDifferentForTest queues two untargeted triggers with
// different sources and labels for owner — a batch the default asks
// about — and returns them in queue order.
func queueTwoDifferentForTest(g *Game, owner *Player, log *[]string) (*StackItem, *StackItem) {
	a := queueTriggerForTest(g, owner, uuid.New(), "A", log)
	b := queueTriggerForTest(g, owner, uuid.New(), "B", log)
	return a, b
}

// drainForTest runs the APNAP drain once and reports whether it drained.
func drainForTest(g *Game) bool {
	var drained bool
	g.WithWriteLock(func() { drained = g.drainPendingTriggersAPNAPLocked() })
	return drained
}

func TestTriggerOrderModeDefaultsToWhenItMatters(t *testing.T) {
	g := newActiveGame(t)
	for _, p := range g.Seats {
		if p.TriggerOrder != TriggerOrderWhenItMatters {
			t.Fatalf("seat %d starts with mode %q", p.Seat, p.TriggerOrder)
		}
	}
}

// Each mode, for a batch of two different untargeted triggers.
func TestTriggerOrderModesOnTwoDifferentTriggers(t *testing.T) {
	for _, c := range []struct {
		mode TriggerOrderMode
		asks bool
	}{
		{TriggerOrderWhenItMatters, true},
		{TriggerOrderAlways, true},
		{TriggerOrderNever, false},
	} {
		g := newActiveGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		if err := g.SetTriggerOrderPreference(me.ID, c.mode); err != nil {
			t.Fatal(err)
		}
		var log []string
		a, b := queueTwoDifferentForTest(g, me, &log)
		drained := drainForTest(g)
		prompt := findTriggerOrderPrompt(g, me.ID)
		if c.asks {
			if drained || prompt == nil || len(prompt.TriggerOrderIDs) != 2 {
				t.Errorf("mode %q: drained=%v prompt=%+v, want a 2-item prompt", c.mode, drained, prompt)
			}
			continue
		}
		if !drained || prompt != nil {
			t.Fatalf("mode %q: drained=%v prompt=%+v, want no prompt", c.mode, drained, prompt)
		}
		// Owner decision 2: the order they triggered in. The first
		// collected goes on the stack first, so it resolves last.
		if a.Seq >= b.Seq {
			t.Errorf("mode %q: placed A at %d and B at %d, want the collected order", c.mode, a.Seq, b.Seq)
		}
	}
}

// Never also covers a batch that the default would hold for a target:
// the seat asked not to be asked.
func TestTriggerOrderNeverSkipsATargetedPair(t *testing.T) {
	items := []*StackItem{
		{ID: uuid.New(), SourceCardID: uuid.New(), Label: "A", Targets: []TargetRef{{Kind: TargetPlayer, ID: uuid.New()}}},
		{ID: uuid.New(), SourceCardID: uuid.New(), Label: "B"},
	}
	if seatNeedsTriggerOrder(items, TriggerOrderNever) {
		t.Error("never asked")
	}
	if !seatNeedsTriggerOrder(items, TriggerOrderWhenItMatters) {
		t.Error("when_it_matters skipped a targeted pair")
	}
}

func TestAlwaysAskPromptsAnAllProwessBatch(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	if err := g.SetTriggerOrderPreference(me.ID, TriggerOrderAlways); err != nil {
		t.Fatal(err)
	}
	queueProwessForTest(g, me)
	queueProwessForTest(g, me)
	queueProwessForTest(g, me)
	if drainForTest(g) {
		t.Error("an always-ask seat's prowess batch drained without a prompt")
	}
	prompt := findTriggerOrderPrompt(g, me.ID)
	if prompt == nil || len(prompt.TriggerOrderIDs) != 3 {
		t.Fatalf("want a 3-item trigger_order prompt, got %+v", prompt)
	}
}

func TestWhenItMattersKeepsAutoOrderingProwess(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	queueProwessForTest(g, me)
	queueProwessForTest(g, me)
	if !drainForTest(g) {
		t.Error("default seat was held behind an ordering prompt")
	}
}

// APNAP across two seats with different modes: each seat's batch is
// decided by its own mode, and the never seat's batch waits for the
// other seat's prompt like any APNAP drain (CR 603.3b).
func TestTriggerOrderModesAreDecidedPerSeat(t *testing.T) {
	g := newActiveGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if err := g.SetTriggerOrderPreference(active.ID, TriggerOrderNever); err != nil {
		t.Fatal(err)
	}
	var log []string
	a1, a2 := queueTwoDifferentForTest(g, active, &log)
	o1 := queueTriggerForTest(g, other, uuid.New(), "O1", &log)
	o2 := queueTriggerForTest(g, other, uuid.New(), "O2", &log)
	if drainForTest(g) {
		t.Fatal("drained while the default seat's prompt is open")
	}
	if findTriggerOrderPrompt(g, active.ID) != nil {
		t.Error("the never seat was asked")
	}
	prompt := findTriggerOrderPrompt(g, other.ID)
	if prompt == nil {
		t.Fatal("the default seat was not asked")
	}
	// O1 resolves first, so it is placed last.
	if err := g.ResolveTriggerOrder(prompt.ID, other.ID, []uuid.UUID{o1.ID, o2.ID}); err != nil {
		t.Fatal(err)
	}
	// Active player's first, in collected order; then the other seat's,
	// in the chosen order.
	if !(a1.Seq < a2.Seq && a2.Seq < o2.Seq && o2.Seq < o1.Seq) {
		t.Errorf("placement a1=%d a2=%d o2=%d o1=%d, want a1<a2<o2<o1", a1.Seq, a2.Seq, o2.Seq, o1.Seq)
	}
}

// An answered batch (every item Ordered) is not asked twice.
func TestAlwaysAskDoesNotRepromptAnOrderedBatch(t *testing.T) {
	items := []*StackItem{{Commutes: true, Ordered: true}, {Commutes: true, Ordered: true}}
	if seatNeedsTriggerOrder(items, TriggerOrderAlways) {
		t.Error("an Ordered batch re-prompted")
	}
	items[1].Ordered = false
	if !seatNeedsTriggerOrder(items, TriggerOrderAlways) {
		t.Error("a batch with an unordered item did not prompt")
	}
}

func TestTriggerOrderModeRoundTripsCloneAndSnapshot(t *testing.T) {
	for _, mode := range []TriggerOrderMode{TriggerOrderAlways, TriggerOrderNever} {
		g := newActiveGame(t)
		if err := g.SetTriggerOrderPreference(g.Seats[1].ID, mode); err != nil {
			t.Fatal(err)
		}
		c := g.Clone()
		if c.Seats[1].TriggerOrder != mode || c.Seats[0].TriggerOrder != TriggerOrderWhenItMatters {
			t.Errorf("%q: Clone lost or smeared the mode", mode)
		}
		raw, err := json.Marshal(g.CaptureSnapshot())
		if err != nil {
			t.Fatal(err)
		}
		var snap GameSnapshot
		if err := json.Unmarshal(raw, &snap); err != nil {
			t.Fatal(err)
		}
		r, err := snap.Restore()
		if err != nil {
			t.Fatal(err)
		}
		if r.Seats[1].TriggerOrder != mode || r.Seats[0].TriggerOrder != TriggerOrderWhenItMatters {
			t.Errorf("%q: snapshot round trip gave %q / %q", mode, r.Seats[1].TriggerOrder, r.Seats[0].TriggerOrder)
		}
	}
}

// A file from before #1968 has only #1530's boolean; true restores as
// always. A mode this binary does not know restores as the default.
func TestTriggerOrderRestoresFromTheOldBoolean(t *testing.T) {
	for _, c := range []struct {
		mode      string
		alwaysAsk bool
		want      TriggerOrderMode
	}{
		{"", true, TriggerOrderAlways},
		{"", false, TriggerOrderWhenItMatters},
		{"never", false, TriggerOrderNever},
		{"always", true, TriggerOrderAlways},
		{"sometimes", true, TriggerOrderMode("sometimes")},
	} {
		if got := restoredTriggerOrder(c.mode, c.alwaysAsk); got != c.want {
			t.Errorf("restoredTriggerOrder(%q, %v) = %q, want %q", c.mode, c.alwaysAsk, got, c.want)
		}
	}
	// An unknown mode behaves as the default.
	var log []string
	g0 := newActiveGame(t)
	pair := []*StackItem{
		queueTriggerForTest(g0, g0.Seats[0], uuid.New(), "A", &log),
		queueTriggerForTest(g0, g0.Seats[0], uuid.New(), "B", &log),
	}
	if !seatNeedsTriggerOrder(pair, TriggerOrderMode("sometimes")) {
		t.Error("an unknown mode skipped a batch the default asks about")
	}
	prowess := []*StackItem{queueProwessForTest(g0, g0.Seats[0]), queueProwessForTest(g0, g0.Seats[0])}
	if seatNeedsTriggerOrder(prowess, TriggerOrderMode("sometimes")) {
		t.Error("an unknown mode asked about a batch the default skips")
	}
	// And through a real file: an "always" seat still writes the old key
	// for a binary from before #1968.
	g := newActiveGame(t)
	if err := g.SetTriggerOrderPreference(g.Seats[0].ID, TriggerOrderAlways); err != nil {
		t.Fatal(err)
	}
	snap := g.CaptureSnapshot()
	if !snap.Seats[0].TriggerOrderAlwaysAsk || snap.Seats[0].TriggerOrder != "always" {
		t.Errorf("snapshot seat = %+v / %q", snap.Seats[0].TriggerOrderAlwaysAsk, snap.Seats[0].TriggerOrder)
	}
}

// Undo is Clone + RestoreFrom. The mode is a setting, set with no undo
// entry, so rewinding an earlier action must not flip it back.
func TestTriggerOrderModeSurvivesUndo(t *testing.T) {
	for _, mode := range []TriggerOrderMode{TriggerOrderAlways, TriggerOrderNever} {
		g := newActiveGame(t)
		pre := g.Clone()
		if err := g.SetTriggerOrderPreference(g.Seats[0].ID, mode); err != nil {
			t.Fatal(err)
		}
		g.WithWriteLock(func() { g.RestoreFrom(pre) })
		if g.Seats[0].TriggerOrder != mode {
			t.Errorf("undo reverted %q to %q", mode, g.Seats[0].TriggerOrder)
		}
	}
}

func TestSetTriggerOrderPreferenceRefusals(t *testing.T) {
	g := NewGame()
	if err := g.SetTriggerOrderPreference(g.ID, TriggerOrderAlways); err == nil {
		t.Error("accepted a preference before the game started")
	}
	g = newActiveGame(t)
	if err := g.SetTriggerOrderPreference(g.ID, TriggerOrderAlways); err == nil {
		t.Error("accepted a preference for a non-seat")
	}
	if err := g.SetTriggerOrderPreference(g.Seats[0].ID, TriggerOrderMode("sometimes")); err == nil {
		t.Error("accepted an unknown mode")
	}
}

func TestParseTriggerOrderMode(t *testing.T) {
	for wire, want := range map[string]TriggerOrderMode{
		"when_it_matters": TriggerOrderWhenItMatters,
		"always":          TriggerOrderAlways,
		"never":           TriggerOrderNever,
	} {
		got, err := ParseTriggerOrderMode(wire)
		if err != nil || got != want {
			t.Errorf("ParseTriggerOrderMode(%q) = %q, %v", wire, got, err)
		}
		if got.Wire() != wire {
			t.Errorf("%q.Wire() = %q", got, got.Wire())
		}
	}
	for _, bad := range []string{"", "Always", "sometimes"} {
		if _, err := ParseTriggerOrderMode(bad); err == nil {
			t.Errorf("ParseTriggerOrderMode(%q) accepted", bad)
		}
	}
}

// --- #1968's widened skip: copies of one source-blind catalog ability.

const copiesGainKey = "Trig Probe — you gain 1 life at upkeep"

// sourceBlindUpkeepRow is a row the catalog registry would mark
// source-blind: no Build, no clause, an effect that reads the
// controller only. Set by hand here, since the registry lives in
// effects (cards/effects/source_blind.go).
func sourceBlindUpkeepRow(blind bool) TriggeredAbility {
	return TriggeredAbility{
		Watches:   []EventKind{EventBeginUpkeep},
		AppliesTo: trigYourUpkeep,
		Key:       copiesGainKey,
		Effect: func(g *Game, item *StackItem) error {
			return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, 1)
		},
		SourceBlind: blind,
	}
}

// fireUpkeepWithProbes puts n probes on seat 0's battlefield and fires
// seat 0's upkeep, returning the queued items in queue order.
func fireUpkeepWithProbes(t *testing.T, g *Game, n int) []*StackItem {
	t.Helper()
	me := g.Seats[0].ID
	for i := 0; i < n; i++ {
		pushTypedTestCard(g, Card{
			Name: "Trig Probe", TypeLine: "Enchantment", OracleID: trigRefOracle,
			Owner: me, Controller: me,
		})
	}
	var items []*StackItem
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventBeginUpkeep, Actor: me})
		items = append(items, g.PendingTriggers...)
	})
	if len(items) != n {
		t.Fatalf("the upkeep queued %d triggers, want %d", len(items), n)
	}
	return items
}

// Two copies of a source-blind ability from two objects (two Soul
// Wardens) do not ask in when_it_matters, and go on the stack in the
// order they were collected.
func TestCopiesOfASourceBlindAbilityDoNotAsk(t *testing.T) {
	withTrigCatalog(t, sourceBlindUpkeepRow(true))
	g := newActiveGame(t)
	items := fireUpkeepWithProbes(t, g, 2)
	if items[0].SourceCardID == items[1].SourceCardID {
		t.Fatal("the probes share a source")
	}
	if !drainForTest(g) {
		t.Fatal("two copies of a source-blind ability were held behind a prompt")
	}
	if findTriggerOrderPrompt(g, g.Seats[0].ID) != nil {
		t.Error("asked to order two copies of a source-blind ability")
	}
	if items[0].Seq >= items[1].Seq {
		t.Errorf("placed %d then %d, want the collected order", items[0].Seq, items[1].Seq)
	}
}

// The same pair from a row that is not source-blind still asks: a
// shared row is the same code, not proof the order cannot matter.
func TestCopiesOfANonSourceBlindAbilityStillAsk(t *testing.T) {
	withTrigCatalog(t, sourceBlindUpkeepRow(false))
	g := newActiveGame(t)
	fireUpkeepWithProbes(t, g, 2)
	if drainForTest(g) {
		t.Fatal("drained two copies of an ability the registry did not mark source-blind")
	}
	if findTriggerOrderPrompt(g, g.Seats[0].ID) == nil {
		t.Error("no prompt")
	}
}

// Always still asks for the copies.
func TestAlwaysAsksForCopiesOfASourceBlindAbility(t *testing.T) {
	withTrigCatalog(t, sourceBlindUpkeepRow(true))
	g := newActiveGame(t)
	if err := g.SetTriggerOrderPreference(g.Seats[0].ID, TriggerOrderAlways); err != nil {
		t.Fatal(err)
	}
	fireUpkeepWithProbes(t, g, 2)
	if drainForTest(g) {
		t.Fatal("an always seat was not asked")
	}
}

// Anything recorded for one copy alone — a target, a mode, a payload,
// an X — or a copy of some other row brings the prompt back.
func TestSourceBlindCopiesNeedNothingChosenPerItem(t *testing.T) {
	withTrigCatalog(t, sourceBlindUpkeepRow(true))
	g := newActiveGame(t)
	items := fireUpkeepWithProbes(t, g, 2)
	if seatNeedsTriggerOrder(items, TriggerOrderWhenItMatters) {
		t.Fatal("the plain pair asks")
	}
	for name, mutate := range map[string]func(it *StackItem){
		"target":  func(it *StackItem) { it.Targets = []TargetRef{{Kind: TargetPlayer, ID: uuid.New()}} },
		"mode":    func(it *StackItem) { it.Modes = []int{0} },
		"payload": func(it *StackItem) { it.Payload = []TargetRef{{Kind: TargetPlayer, ID: uuid.New()}} },
		"x":       func(it *StackItem) { it.XValue = 2 },
		"another row": func(it *StackItem) {
			ref := *it.Params.Ability
			ref.Ref = "own:7"
			it.Params.Ability = &ref
		},
		"unstamped": func(it *StackItem) { it.Body = "" },
	} {
		second := *items[1]
		mutate(&second)
		if !seatNeedsTriggerOrder([]*StackItem{items[0], &second}, TriggerOrderWhenItMatters) {
			t.Errorf("%s: a copy with something of its own did not ask", name)
		}
	}
	// One copy plus a different trigger asks.
	other := &StackItem{ID: uuid.New(), SourceCardID: uuid.New(), Label: "Other"}
	if !seatNeedsTriggerOrder([]*StackItem{items[0], items[1], other}, TriggerOrderWhenItMatters) {
		t.Error("copies plus a different trigger did not ask")
	}
}
