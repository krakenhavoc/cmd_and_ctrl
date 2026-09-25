package effects

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// target_set_test.go — #1559: a rule over the chosen SET of a clause's
// targets (CR 601.2c), enforced at announce and re-judged at
// resolution (CR 608.2b), and the "mana value X or less" bound that
// Agadeem's Awakening prints beside it. See game/target_set.go and ADR
// 0019's 2026-09-24 amendment.

const (
	oracleAgadeemsAwakening  = "562d71b9-1646-474e-9293-55da6947a758"
	oracleBeholdSinisterSix  = "6fa27f8a-bade-460f-853a-abc1c05c946a"
	oracleMoltenPrimordial   = "8d8c9f7b-92c7-4284-ad9c-304ce42edba5"
	oracleRunAwayTogether    = "290faa28-450e-4797-9a8f-642d8af3f82a"
	oracleWindgracesJudgment = "de1ca6ed-b275-4f62-ba05-f31b3659b352"
)

// graveCreature puts a creature card with a real mana cost into p's
// graveyard — the mana value is the point of every test below.
func graveCreature(p *game.Player, name, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Zombie", ManaCost: cost,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// castX casts a catalog spell from the active seat's hand announcing
// x, and returns the announce error rather than failing on it.
func castX(t *testing.T, g *game.Game, name, typeLine, oracle string, x int, targets []game.TargetRef) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: active.ID, Controller: active.ID,
	})
	advanceToMain(t, g)
	return g.CastSpell(active.ID, id, game.CastSpellParams{XValue: x, Targets: targets})
}

func TestTargetSetProofCardsAreRegistered(t *testing.T) {
	for _, tc := range []struct {
		oracle, name string
		full         bool
	}{
		{oracleAgadeemsAwakening, "Agadeem's Awakening", true},
		{oracleBeholdSinisterSix, "Behold the Sinister Six!", true},
		{oracleMoltenPrimordial, "Molten Primordial", true},
		{oracleRunAwayTogether, "Run Away Together", true},
		{oracleWindgracesJudgment, "Windgrace's Judgment", false},
	} {
		spec, ok := Lookup(tc.oracle)
		if !ok || spec.Name != tc.name {
			t.Errorf("%s: registered = %v as %q", tc.name, ok, spec.Name)
			continue
		}
		if full := spec.Completeness == CompletenessFull; full != tc.full {
			t.Errorf("%s: Completeness = %s, want full=%v", tc.name, spec.Completeness, tc.full)
		}
	}
}

// Agadeem's Awakening: a legal set of different mana values, each X or
// less, returns; two of one mana value, or one above X, is refused at
// announce with the reason.
func TestAgadeemsAwakeningReturnsDifferentManaValuesXOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	one := graveCreature(me, "One Drop", "{B}")
	twoA := graveCreature(me, "Two Drop A", "{1}{B}")
	twoB := graveCreature(me, "Two Drop B", "{B}{B}")
	three := graveCreature(me, "Three Drop", "{2}{B}")
	zero := graveCreature(me, "Zero Drop", "")

	err := castX(t, g, "Agadeem's Awakening", "Sorcery", oracleAgadeemsAwakening, 2, cardRefs(one, twoA, twoB))
	if !errors.Is(err, game.ErrIllegalTarget) || !strings.Contains(err.Error(), "different mana value") {
		t.Fatalf("two cards of mana value 2 are refused naming the rule, got %v", err)
	}
	err = castX(t, g, "Agadeem's Awakening", "Sorcery", oracleAgadeemsAwakening, 2, cardRefs(one, three))
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("mana value 3 is above X=2 and is refused, got %v", err)
	}
	if err := castX(t, g, "Agadeem's Awakening", "Sorcery", oracleAgadeemsAwakening, 2, cardRefs(zero, one, twoA)); err != nil {
		t.Fatalf("mana values 0, 1 and 2 at X=2 are a legal set: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{zero, one, twoA} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s should have returned", id)
		}
	}
	for _, id := range []uuid.UUID{twoB, three} {
		if b12ZoneOf(g, id) != game.ZoneGraveyard {
			t.Errorf("%s was not picked and stays", id)
		}
	}
	// "Any number" includes none.
	if err := castX(t, g, "Agadeem's Awakening", "Sorcery", oracleAgadeemsAwakening, 0, nil); err != nil {
		t.Fatalf("X=0 with no targets is a legal cast: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// The hand snapshot's legal set is built before X is known, so it is a
// superset — the X bound is the client's to apply from the shipped mana
// values. Every creature card in your graveyard is in it.
func TestAgadeemsAwakeningLegalSetBeforeXIsASuperset(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cheap := graveCreature(me, "Cheap", "{B}")
	dear := graveCreature(me, "Dear", "{6}{B}")
	legal := legalCards(g, me.ID, oracleAgadeemsAwakening)
	if !legal[cheap] || !legal[dear] {
		t.Errorf("before X is announced every creature card qualifies, got %v", legal)
	}
}

// CR 608.2b: a target that left in response is dropped alone and the
// rest still return.
func TestAgadeemsAwakeningOneTargetLeavingCostsOnlyThatCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	one := graveCreature(me, "One Drop", "{B}")
	two := graveCreature(me, "Two Drop", "{1}{B}")
	if err := castX(t, g, "Agadeem's Awakening", "Sorcery", oracleAgadeemsAwakening, 3, cardRefs(one, two)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(one) })
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(two) {
		t.Error("the surviving target still returns")
	}
	if g.Battlefield.Contains(one) {
		t.Error("the exiled target does not")
	}
}

// Run Away Together, re-judged at resolution: a control change in
// response that leaves both creatures under ONE player breaks the pair,
// and neither is a legal target — the spell does nothing. See
// setRuleConflictLocked for why neither is preferred.
func TestRunAwayTogetherPairBrokenInResponseReturnsNeither(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine, theirs := seedCreature(g, "My Bear", me.ID), seedCreature(g, "Their Bear", opp.ID)
	castCatalogSpell(t, g, "Run Away Together", "Instant", oracleRunAwayTogether, cardRefs(mine, theirs))
	// Their Bear comes to me: both targets now share a controller.
	stealForTest(t, g, theirs, me.ID)
	if c, _ := battlefieldCard(g, theirs); c.Controller != me.ID {
		t.Fatalf("setup: the steal should hand me Their Bear, controller = %s", c.Controller)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(theirs) {
		t.Error("a pair under one controller breaks the rule — neither is returned")
	}
}

// Behold the Sinister Six!: up to six, different names.
func TestBeholdTheSinisterSixReturnsDifferentNames(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var six []uuid.UUID
	for _, n := range []string{"Doc Ock", "Vulture", "Electro", "Kraven", "Mysterio", "Sandman"} {
		six = append(six, graveCreature(me, n, "{2}{B}"))
	}
	twin := graveCreature(me, "Vulture", "{2}{B}")
	seventh := graveCreature(me, "Rhino", "{3}{B}")

	err := castCatalogSpellErr(t, g, "Behold the Sinister Six!", "Sorcery", oracleBeholdSinisterSix, cardRefs(six[0], six[1], twin))
	if !errors.Is(err, game.ErrIllegalTarget) || !strings.Contains(err.Error(), "different name") {
		t.Fatalf("two Vultures are refused naming the rule, got %v", err)
	}
	if err := castCatalogSpellErr(t, g, "Behold the Sinister Six!", "Sorcery", oracleBeholdSinisterSix,
		cardRefs(append(append([]uuid.UUID(nil), six...), seventh)...)); err == nil {
		t.Fatal("seven targets is one more than up to six")
	}
	castCatalogSpell(t, g, "Behold the Sinister Six!", "Sorcery", oracleBeholdSinisterSix, cardRefs(six...))
	passPriorityAroundTable(t, g)
	for _, id := range six {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s should have returned", id)
		}
	}
	if b12ZoneOf(g, twin) != game.ZoneGraveyard || b12ZoneOf(g, seventh) != game.ZoneGraveyard {
		t.Error("the cards not picked stay")
	}
}

// Molten Primordial: one clause per opponent, each binding its pick to
// that player (the triage on #1559). The trigger asks once per opponent
// and takes one creature from each.
func TestMoltenPrimordialTakesUpToOneCreatureFromEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var opps []*game.Player
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opps = append(opps, p)
		}
	}
	a := pushCreatureToBattlefieldForTest(g, opps[0].ID, "A's Bear")
	a2 := pushCreatureToBattlefieldForTest(g, opps[0].ID, "A's Other Bear")
	b := pushCreatureToBattlefieldForTest(g, opps[1].ID, "B's Bear")
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(a) })
	castHoldCreature(t, g, "Molten Primordial", "Creature — Avatar", oracleMoltenPrimordial, 6, 4)

	// Opponent one: only their creatures are offered, and one pick.
	ch := pickTargetPrompt(t, g)
	if !slices.Contains(ch.PickTargetCards, a) || slices.Contains(ch.PickTargetCards, b) || ch.PickTargetMax != 1 || ch.PickTargetMin != 0 {
		t.Fatalf("first clause offers only the first opponent's creatures, up to one: %v (%d..%d)", ch.PickTargetCards, ch.PickTargetMin, ch.PickTargetMax)
	}
	if !strings.Contains(ch.Reason, opps[0].Name) {
		t.Errorf("the prompt names the opponent it is asking about: %q", ch.Reason)
	}
	answerPickTarget(t, g, a)
	answerPickTarget(t, g, b)
	// Opponent three controls nothing; their clause is skipped.
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{a, b} {
		c, ok := battlefieldCard(g, id)
		if !ok || c.Controller != me.ID {
			t.Errorf("%s is mine until end of turn", id)
		}
		if c.Tapped {
			t.Errorf("%s is untapped", id)
		}
		if !slices.Contains(c.Effective().Abilities, "haste") {
			t.Errorf("%s has haste", id)
		}
	}
	if c, _ := battlefieldCard(g, a2); c.Controller != opps[0].ID {
		t.Error("only one creature per opponent")
	}
}

// The binding is what distinguishes Molten Primordial from a set rule:
// a creature that changes hands in response is not controlled by THE
// PLAYER its clause names, so it is an illegal target — even when its
// new controller is another opponent.
func TestMoltenPrimordialPickThatChangedHandsIsNotTaken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var opps []*game.Player
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opps = append(opps, p)
		}
	}
	a := pushCreatureToBattlefieldForTest(g, opps[0].ID, "A's Bear")
	b := pushCreatureToBattlefieldForTest(g, opps[1].ID, "B's Bear")
	castHoldCreature(t, g, "Molten Primordial", "Creature — Avatar", oracleMoltenPrimordial, 6, 4)
	answerPickTarget(t, g, a)
	answerPickTarget(t, g, b)
	// The two creatures swap controllers in response.
	stealForTest(t, g, a, opps[1].ID)
	stealForTest(t, g, b, opps[0].ID)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, a); c.Controller != opps[1].ID {
		t.Error("A's Bear now belongs to the second opponent and is not taken")
	}
	if c, _ := battlefieldCard(g, b); c.Controller != opps[0].ID {
		t.Error("B's Bear now belongs to the first opponent and is not taken")
	}
}

// The view and the gate agree: the hand card's legal_targets carries
// the set rule's keys, and every key the gate would refuse to repeat is
// a key the view ships.
func TestSetRuleKeysMatchTheGate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	one := graveCreature(me, "One", "{B}")
	twoA := graveCreature(me, "Two A", "{1}{B}")
	twoB := graveCreature(me, "Two B", "{B}{B}")
	spec := game.TargetSpecFor(oracleAgadeemsAwakening)
	var keys map[uuid.UUID]string
	g.WithWriteLock(func() { keys = g.TargetDifferenceKeysForEffect(spec, []uuid.UUID{one, twoA, twoB}) })
	if keys[one] == keys[twoA] || keys[twoA] != keys[twoB] {
		t.Fatalf("keys = %v: the two mana-value-2 cards share one, the 1-drop differs", keys)
	}
}
