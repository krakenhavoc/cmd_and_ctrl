package game

import (
	"testing"

	"github.com/google/uuid"
)

// produce_mana_test.go — #1222, the engine half of the mana-production
// window. The catalog half (the real Mana Reflection and Nyxbloom
// Ancient) is in cards/effects/mana_replacements_test.go, where a card
// can carry the replacement and be scoped to its controller.
//
// What is here is the engine's own notion of a PRODUCTION: one CR 614
// window per batch of mana about to reach a pool, opened at all four
// production sites, carrying the colours rather than a bare count, and
// never able to pause (CR 605.3a).

// manaProducedReplacement is "if <owner> taps a permanent for mana, it
// produces `times` as much of that mana instead" as a test injection.
func manaProducedReplacement(owner uuid.UUID, times int, label string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventManaAdded},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventProduceMana && ev.ManaFromTap && len(ev.ManaColors) > 0 &&
				ev.ManaPlayer == owner
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.MultiplyMana(times)
			return nil
		},
		Controller: func(*ReplacementEvent, *Game, *Card) uuid.UUID { return owner },
		Label:      label,
	}
}

// manaProducedCancelReplacement is the CR 614.10 null replacement: the
// production simply does not happen.
func manaProducedCancelReplacement(label string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventManaAdded},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventProduceMana
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.Cancel()
			return nil
		},
		Label: label,
	}
}

// poolList is the pool as a colour list, in the order it arrived —
// the sibling of autotap_one_color_test.go's poolColors tally, which
// answers "how many of each" rather than "in what order".
func poolList(p *Player) []string {
	out := make([]string, 0, len(p.ManaPool))
	for _, t := range p.ManaPool {
		out = append(out, t.Color)
	}
	return out
}

func wantPool(t *testing.T, p *Player, want ...string) {
	t.Helper()
	got := poolList(p)
	if len(got) != len(want) {
		t.Fatalf("pool: got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("pool: got %v, want %v", got, want)
		}
	}
}

// TestManaProductionDoubledAtTheActivation is the first of the four
// production sites: a printed single-colour slot, doubled before it is
// in the pool.
func TestManaProductionDoubledAtTheActivation(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	forest := pushBattlefieldForTest(g, p.ID, "Forest", "Basic Land — Forest", "")
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(manaProducedReplacement(p.ID, 2, "double"))
	})

	if err := g.ActivateManaAbility(p.ID, forest, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	wantPool(t, p, "G", "G")
}

// TestManaProductionDoubledPerSlot pins "twice as much of THAT mana"
// on a two-slot ability: a Sol Ring makes four, not two.
func TestManaProductionDoubledPerSlot(t *testing.T) {
	withCatalogHook(t, solRingHook)
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	ring := pushBattlefieldForTest(g, p.ID, "Sol Ring", "Artifact", "6ad8011d-3471-4369-9d68-b264cc027487")
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(manaProducedReplacement(p.ID, 2, "double"))
	})

	if err := g.ActivateManaAbility(p.ID, ring, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	wantPool(t, p, "C", "C", "C", "C")
}

// TestManaProductionDoubledAtThePick is the second production site,
// and the one that only exists because the colour is not known any
// earlier: a Birds of Paradise is ONE pick that mints two of the
// chosen colour.
func TestManaProductionDoubledAtThePick(t *testing.T) {
	withCatalogHook(t, birdsHook)
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	birds := pushBattlefieldForTest(g, p.ID, "Birds of Paradise", "Creature — Bird", "d3a0b660-358c-41bd-9cd2-41fbf3491b1a")
	// CR 302.6: it has to have been under their control since the turn
	// began, which pushBattlefieldForTest does not arrange.
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == birds {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	}
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(manaProducedReplacement(p.ID, 2, "double"))
	})

	if err := g.ActivateManaAbility(p.ID, birds, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(p.ManaPool) != 0 {
		t.Fatalf("the pick is unanswered, so nothing should be in the pool yet: %v", poolList(p))
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != PendingChoiceMana {
		t.Fatalf("expected one mana pick, got %+v", g.PendingChoices)
	}
	if err := g.ResolveManaChoice(g.PendingChoices[0].ID, p.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	// The SAME colour twice. CR 106.12b replaces how much mana is
	// produced, not how many choices the player gets.
	wantPool(t, p, "U", "U")
}

// TestManaProductionNotDoubledForASpell is the printed condition: "if
// you TAP a permanent for mana". A resolving spell's "Add {B}{B}{B}"
// taps nothing.
func TestManaProductionNotDoubledForASpell(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(manaProducedReplacement(p.ID, 2, "double"))
		if err := g.AddManaForEffect(p.ID, uuid.New(), "{B}{B}{B}"); err != nil {
			t.Fatalf("AddManaForEffect: %v", err)
		}
	})
	wantPool(t, p, "B", "B", "B")
}

// TestTwoManaDoublersCompose is the CR 616.1 apply-loop: one
// replacement, then re-gather, so a ×2 and a ×3 in one window are ×6.
// No prompt is queued for the ordering — a production cannot pause —
// and none is needed, because multiplication commutes.
func TestTwoManaDoublersCompose(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	forest := pushBattlefieldForTest(g, p.ID, "Forest", "Basic Land — Forest", "")
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(manaProducedReplacement(p.ID, 2, "double"))
		g.RegisterReplacementForTest(manaProducedReplacement(p.ID, 3, "triple"))
	})

	if err := g.ActivateManaAbility(p.ID, forest, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(p.ManaPool) != 6 {
		t.Fatalf("x2 then x3 on one {G}: got %v, want six", poolList(p))
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceReplacementOrder {
			t.Fatalf("a mana production must never queue a CR 616 ordering prompt (CR 605.3a)")
		}
	}
}

// TestManaProductionCanBeReplacedAway is CR 614.10's null
// replacement. The land is still tapped — the cost was paid — and the
// pool is empty.
func TestManaProductionCanBeReplacedAway(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	forest := pushBattlefieldForTest(g, p.ID, "Forest", "Basic Land — Forest", "")
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(manaProducedCancelReplacement("no mana"))
	})

	if err := g.ActivateManaAbility(p.ID, forest, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(p.ManaPool) != 0 {
		t.Errorf("cancelled production: pool %v, want empty", poolList(p))
	}
	if !cardTappedForTest(g, forest) {
		t.Errorf("the cost was paid, so the Forest is tapped whatever the window did")
	}
}

// TestManaProductionDoublesTheProducedRestrictions keeps #259 closed
// across the new seam: a doubled restricted source must not put
// unrestricted mana in the pool.
func TestManaProductionDoublesTheProducedRestrictions(t *testing.T) {
	withCatalogHook(t, func(oracleID string) []ManaAbilityShape {
		if oracleID == "ziggurat" {
			return []ManaAbilityShape{{
				TapCost:      true,
				Produced:     "{G}",
				Restrictions: []string{ManaRestrictType("Creature")},
				Label:        "Add {G}",
			}}
		}
		return nil
	})
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	zig := pushBattlefieldForTest(g, p.ID, "Ancient Ziggurat", "Land", "ziggurat")
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(manaProducedReplacement(p.ID, 2, "double"))
	})

	if err := g.ActivateManaAbility(p.ID, zig, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(p.ManaPool) != 2 {
		t.Fatalf("doubled Ziggurat: got %v, want two", poolList(p))
	}
	for i, tok := range p.ManaPool {
		if len(tok.Restrictions) != 1 {
			t.Errorf("token %d: restrictions %v, want the ability's", i, tok.Restrictions)
		}
	}
}

// TestManaProductionKeepsTheSourceKindsSnapshot is the seam with
// #1212: CR 106.12b replaces how much mana is produced, not what
// produced it, so every mana the window ADDS carries the same snapshot
// the printed one would have. A doubled Treasure is two Treasure mana.
//
// It matters because collapsing the four mint sites into one body
// moved the STAMPING and left the SNAPSHOTTING per-site; this is the
// assertion that the move did not drop it.
func TestManaProductionKeepsTheSourceKindsSnapshot(t *testing.T) {
	withCatalogHook(t, func(oracleID string) []ManaAbilityShape {
		if oracleID == "treasure-token" {
			return []ManaAbilityShape{{TapCost: true, Produced: "{G}", Label: "Add {G}"}}
		}
		return nil
	})
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	treasure := pushBattlefieldForTest(g, p.ID, "Treasure", "Artifact — Treasure", "treasure-token")
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(manaProducedReplacement(p.ID, 2, "double"))
	})

	if err := g.ActivateManaAbility(p.ID, treasure, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(p.ManaPool) != 2 {
		t.Fatalf("a doubled Treasure: got %v, want two", poolList(p))
	}
	for i, tok := range p.ManaPool {
		if !tok.SourceKinds.Has(ManaSourceTreasure) {
			t.Errorf("token %d: SourceKinds %b, want the Treasure bit", i, tok.SourceKinds)
		}
		if !tok.SourceKinds.Has(ManaSourceArtifact) {
			t.Errorf("token %d: SourceKinds %b, want the Artifact bit too", i, tok.SourceKinds)
		}
	}
}

// TestManaProductionPlannerPricesTheReplacedAmount is the auto-tap
// PLANNER: a Mana-Reflected Forest pays {G}{G} on its own, so the
// solver may plan one land where it would otherwise need two.
func TestManaProductionPlannerPricesTheReplacedAmount(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	forest := pushBattlefieldForTest(g, p.ID, "Forest", "Basic Land — Forest", "")

	if _, ok := g.AutoTapForCost(p.ID, costFor(t, "{G}{G}"), 0); ok {
		t.Fatalf("one Forest must not pay {G}{G} without a doubler")
	}
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(manaProducedReplacement(p.ID, 2, "double"))
	})
	plan, ok := g.AutoTapForCost(p.ID, costFor(t, "{G}{G}"), 0)
	if !ok {
		t.Fatalf("a doubled Forest pays {G}{G}")
	}
	if len(plan) != 1 || plan[0] != forest {
		t.Fatalf("plan: got %v, want just the Forest", plan)
	}
}

// TestManaProductionExecutorPaysWithTheReplacedAmount is the other
// half of the same rule: the plan the solver made has to be payable
// when the executor runs it, with no stranded tap and no leftover
// requirement.
func TestManaProductionExecutorPaysWithTheReplacedAmount(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	pushBattlefieldForTest(g, p.ID, "Forest", "Basic Land — Forest", "")
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(manaProducedReplacement(p.ID, 2, "double"))
	})
	id := pushTypedCardToHandWithCost(p, "Nature's Lore", "Sorcery", "{G}{G}")

	if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tap cast of {G}{G} off one doubled Forest: %v (pool %v)", err, poolList(p))
	}
	if len(p.ManaPool) != 0 {
		t.Errorf("post-spend pool: got %v, want empty", poolList(p))
	}
}

// TestManaProductionDoubledPickIsStillOnePick is the planner's side of
// "one pick, two mana of the SAME colour": a doubled Birds of Paradise
// cannot pay {W}{U}, and the planner must not plan as if it could.
func TestManaProductionDoubledPickIsStillOnePick(t *testing.T) {
	withCatalogHook(t, birdsHook)
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	birds := pushBattlefieldForTest(g, p.ID, "Birds of Paradise", "Creature — Bird", "d3a0b660-358c-41bd-9cd2-41fbf3491b1a")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == birds {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	}
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(manaProducedReplacement(p.ID, 2, "double"))
	})

	if _, ok := g.AutoTapForCost(p.ID, costFor(t, "{W}{U}"), 0); ok {
		t.Errorf("a doubled Birds is one pick — it cannot pay two different colours")
	}
	if _, ok := g.AutoTapForCost(p.ID, costFor(t, "{U}{U}"), 0); !ok {
		t.Errorf("a doubled Birds pays {U}{U} with one pick")
	}
	// And the executor really pays it: the plan booked a colour for a
	// slot the PRINTED list does not show as a one-colour pick, and
	// materializePlanLocked honours that booking (plannedTap.OneColor)
	// rather than re-deriving it.
	id := pushTypedCardToHandWithCost(p, "Counterspell", "Instant", "{U}{U}")
	if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tap cast of {U}{U} off one doubled Birds: %v (pool %v)", err, poolList(p))
	}
	if len(p.ManaPool) != 0 {
		t.Errorf("post-spend pool: got %v, want empty", poolList(p))
	}
}
