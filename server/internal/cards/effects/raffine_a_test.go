package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// raffine_a_test.go — the Raffine Reanimator deck, slice raffine-a (#3212).

const raffineOracle = "9307f7e9-bcd9-41f9-88f8-0fb9700a6b12"

func TestRaffineIsRegisteredAndWhole(t *testing.T) {
	spec, ok := Lookup(raffineOracle)
	if !ok {
		t.Fatal("Raffine is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Completeness = %v, Caveats = %v, want Full and none", spec.Completeness, spec.Caveats)
	}
}

// Two attackers: the chosen one connives 2. Draw two, discard two; the nonland
// card is a counter, the land is not.
func TestRaffineConnivesXWhereXIsAttackingCreatures(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Raffine, Scheming Seer", "Legendary Creature — Sphinx Demon", raffineOracle, false)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	nonland := handCardFull(me, "Spell A", "Sorcery", "{1}", "", nil)
	land := handCardFull(me, "Land B", "Basic Land — Forest", "", "", nil)

	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("attack: %v", err)
		}
	}
	lockInAttacks(t, g)
	answerPickTarget(t, g, a)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("a second target prompt: the trigger fires once per attack, not per attacker")
	}
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != 4 {
		t.Fatalf("hand after drawing X = 2: %d, want 4", got)
	}
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no discard prompt")
	}
	// A drawn library card is a second discard; the library fixture is nonland
	// or land, so only assert on the two seeded cards plus the counter count
	// derived from what is actually discarded.
	var drawn []uuid.UUID
	for _, c := range me.Hand.Cards {
		if c.InstanceID != nonland && c.InstanceID != land {
			drawn = append(drawn, c.InstanceID)
		}
	}
	if len(drawn) != 2 {
		t.Fatalf("drawn cards = %d, want 2", len(drawn))
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{nonland, land}); err != nil {
		t.Fatalf("discard: %v", err)
	}
	if got := counterCount(g, a, game.CounterPlusOne); got != 1 {
		t.Errorf("counters on the target = %d, want 1 (one nonland discard)", got)
	}
	if got := counterCount(g, b, game.CounterPlusOne); got != 0 {
		t.Errorf("the other attacker got %d counters", got)
	}
	if me.Hand.Size() != 2 {
		t.Errorf("hand = %d, want 2", me.Hand.Size())
	}
	// CR 701.50f, #4641: the target connived once, however many cards X was.
	if evs := eventsOfKind(g, game.EventConnived); len(evs) != 1 || evs[0].CardID != a || evs[0].Actor != me.ID {
		t.Errorf("connive events = %+v, want one for the target under its controller", evs)
	}
}

// Raffine is not an attacker herself and does not trigger when an opponent
// attacks: "whenever YOU attack".
func TestRaffineDoesNotTriggerOnAnOpponentsAttack(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, opp.ID, "Raffine, Scheming Seer", "Legendary Creature — Sphinx Demon", raffineOracle, false)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	lockInAttacks(t, g)
	if latestPickTarget(g, opp.ID) != nil {
		t.Error("Raffine asked for a target on another player's attack")
	}
}
