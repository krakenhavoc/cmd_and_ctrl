package game

import (
	"testing"

	"github.com/google/uuid"
)

// identical_replacements_test.go covers #792: two objects
// contributing the SAME declared replacement effect are not a CR 616
// ordering question, because every order applies the same
// modification the same number of times. Two Doubling Seasons, two
// Rhox Faithmenders, two Hardened Scales.
//
// The catalog-facing half of this lives in
// cards/effects/doubling_season_test.go and
// cards/effects/rhox_faithmender_test.go, on the real cards. What is
// here is the engine's own notion of "the same effect" —
// replacementIdentity — and the boundaries it draws.

// stubCatalogReplacements points CatalogReplacements at a fixed
// key → effects map for the duration of one test.
func stubCatalogReplacements(t *testing.T, byKey map[string][]ReplacementEffect) {
	t.Helper()
	prev := CatalogReplacements
	CatalogReplacements = func(key string) []ReplacementEffect { return byKey[key] }
	t.Cleanup(func() { CatalogReplacements = prev })
}

// counterDoubler is the shape both halves of these tests use: a
// Doubling-Season-style replacement that doubles any counter placed
// on a permanent its own controller controls.
func counterDoubler(label string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventCounterPlaced},
		AppliesTo: func(ev *ReplacementEvent, g *Game, src *Card) bool {
			if ev.Kind != RepEventCounter || src == nil {
				return false
			}
			target, ok := g.LookupCardForEffect(ev.CounterTarget)
			return ok && target.Controller == src.Controller
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.CounterDelta *= 2
			return nil
		},
		Controller: func(_ *ReplacementEvent, _ *Game, src *Card) uuid.UUID { return src.Controller },
		Label:      label,
	}
}

// pushReplacementSource puts a permanent carrying oracleID's catalog
// replacements onto the battlefield under controller.
func pushReplacementSource(g *Game, oracleID string, controller uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       "Source",
		TypeLine:   "Enchantment",
		OracleID:   oracleID,
		Owner:      controller,
		Controller: controller,
	})
	return id
}

// pushBear puts a plain creature on the battlefield to receive
// counters.
func pushBear(g *Game, controller uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       "Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      controller,
		Controller: controller,
	})
	return id
}

func countersOnCard(g *Game, cardID uuid.UUID, name string) int {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == cardID {
			return c.Counters[name]
		}
	}
	return 0
}

// TestTwoObjectsOneEffectApplyWithoutAPrompt — the headline. Two
// permanents printed with the same replacement double the counter
// twice, inline, and the affected player is asked nothing.
func TestTwoObjectsOneEffectApplyWithoutAPrompt(t *testing.T) {
	const doublerOracle = "00000000-0000-4000-8000-0000000d0071"
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		doublerOracle: {counterDoubler("Doubler")},
	})
	g := newActiveGame(t)
	owner := g.Seats[0].ID
	pushReplacementSource(g, doublerOracle, owner)
	pushReplacementSource(g, doublerOracle, owner)
	bear := pushBear(g, owner)

	if err := g.AddCounter(bear, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("pending choices = %d, want 0 (one effect on two objects)", len(g.PendingChoices))
	}
	if got := countersOnCard(g, bear, "+1/+1"); got != 4 {
		t.Errorf("counters = %d, want 4 (1 × 2 × 2, both copies applied)", got)
	}
}

// TestTwoDifferentEffectsStillPrompt — the guard on the above. Two
// distinct catalog entries are two effects however alike their
// arithmetic looks, and CR 616 says the affected player chooses.
func TestTwoDifferentEffectsStillPrompt(t *testing.T) {
	const (
		firstOracle  = "00000000-0000-4000-8000-0000000d0072"
		secondOracle = "00000000-0000-4000-8000-0000000d0073"
	)
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		firstOracle:  {counterDoubler("First doubler")},
		secondOracle: {counterDoubler("Second doubler")},
	})
	g := newActiveGame(t)
	owner := g.Seats[0].ID
	pushReplacementSource(g, firstOracle, owner)
	pushReplacementSource(g, secondOracle, owner)
	bear := pushBear(g, owner)

	if err := g.AddCounter(bear, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want 1 CR 616 ordering prompt", len(g.PendingChoices))
	}
	if got := countersOnCard(g, bear, "+1/+1"); got != 0 {
		t.Errorf("counters = %d, want 0 while the prompt is unanswered", got)
	}
}

// TestSameCardUnderTwoControllersStillPrompts — the controller half
// of the identity. A replacement routinely reads its own controller
// (Notion Thief writes it straight into the event), so two copies of
// one card under different controllers are two different
// modifications and the order between them is observable.
func TestSameCardUnderTwoControllersStillPrompts(t *testing.T) {
	const doublerOracle = "00000000-0000-4000-8000-0000000d0074"
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		doublerOracle: {{
			Watches: []EventKind{EventCounterPlaced},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, src *Card) bool {
				return ev.Kind == RepEventCounter && src != nil
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.CounterDelta *= 2
				return nil
			},
			Controller: func(_ *ReplacementEvent, _ *Game, src *Card) uuid.UUID { return src.Controller },
			Label:      "Everyone's doubler",
		}},
	})
	g := newActiveGame(t)
	mine, theirs := g.Seats[0].ID, g.Seats[1].ID
	pushReplacementSource(g, doublerOracle, mine)
	pushReplacementSource(g, doublerOracle, theirs)
	bear := pushBear(g, mine)

	if err := g.AddCounter(bear, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want 1 (different controllers)", len(g.PendingChoices))
	}
}

// TestTwoSlotsOfOneCardStillPrompt — the slot half. One card can
// print two different replacements; they are not each other however
// many copies of the card are out.
func TestTwoSlotsOfOneCardStillPrompt(t *testing.T) {
	const twoSlotOracle = "00000000-0000-4000-8000-0000000d0075"
	adder := counterDoubler("Adder")
	adder.Replace = func(ev *ReplacementEvent, _ *Game, _ *Card) error {
		ev.CounterDelta++
		return nil
	}
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		twoSlotOracle: {counterDoubler("Doubler"), adder},
	})
	g := newActiveGame(t)
	owner := g.Seats[0].ID
	pushReplacementSource(g, twoSlotOracle, owner)
	bear := pushBear(g, owner)

	if err := g.AddCounter(bear, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want 1 (two slots on one card)", len(g.PendingChoices))
	}
}

// TestTestReplacementsNeverCollapse — built-in, turn-scoped and
// test-injected replacements are registered per instance rather than
// declared on a catalog entry, so two of them are two declarations
// that happen to look alike. They keep prompting, which is what the
// pre-#792 engine did for everything.
func TestTestReplacementsNeverCollapse(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0].ID
	bear := pushBear(g, owner)

	g.mu.Lock()
	for range 2 {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventCounterPlaced},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventCounter
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.CounterDelta *= 2
				return nil
			},
			Label: "Doubler",
		})
	}
	g.mu.Unlock()

	if err := g.AddCounter(bear, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want 1 (no catalog identity to collapse)", len(g.PendingChoices))
	}
}

// TestSameModificationBoundaries pins the predicate itself, including
// the cases the pipeline has no card to reach yet.
func TestSameModificationBoundaries(t *testing.T) {
	alice, bob := uuid.New(), uuid.New()
	base := replacementIdentity{card: "season", slot: 0, controller: alice}
	plain := ReplacementEffect{Label: "Doubling Season"}

	with := func(id replacementIdentity, e ReplacementEffect) activeReplacement {
		return activeReplacement{effect: e, identity: id}
	}

	cases := []struct {
		name string
		in   []activeReplacement
		want bool
	}{
		{"two objects, one effect", []activeReplacement{with(base, plain), with(base, plain)}, true},
		{"three objects, one effect", []activeReplacement{with(base, plain), with(base, plain), with(base, plain)}, true},
		{"one object", []activeReplacement{with(base, plain)}, false},
		{"none", nil, false},
		{"different card", []activeReplacement{
			with(base, plain),
			with(replacementIdentity{card: "scales", slot: 0, controller: alice}, plain),
		}, false},
		{"different slot", []activeReplacement{
			with(base, plain),
			with(replacementIdentity{card: "season", slot: 1, controller: alice}, plain),
		}, false},
		{"different controller", []activeReplacement{
			with(base, plain),
			with(replacementIdentity{card: "season", slot: 0, controller: bob}, plain),
		}, false},
		{"unidentified", []activeReplacement{
			with(replacementIdentity{}, plain),
			with(replacementIdentity{}, plain),
		}, false},
		{"one identified, one not", []activeReplacement{with(base, plain), with(replacementIdentity{}, plain)}, false},
		{"CR 614.10 may", []activeReplacement{
			with(base, ReplacementEffect{Optional: true}),
			with(base, ReplacementEffect{Optional: true}),
		}, false},
		{"pay-life entry", []activeReplacement{
			with(base, ReplacementEffect{EntryLifeCost: 2}),
			with(base, ReplacementEffect{EntryLifeCost: 2}),
		}, false},
		{"copy selector", []activeReplacement{
			with(base, ReplacementEffect{CopySelector: &CopySelector{}}),
			with(base, ReplacementEffect{CopySelector: &CopySelector{}}),
		}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := sameModification(tc.in); got != tc.want {
				t.Errorf("sameModification = %v, want %v", got, tc.want)
			}
		})
	}
}
