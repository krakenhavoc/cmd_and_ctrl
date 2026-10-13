package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cav_b_connive_test.go — Lethal Scheme's "each creature that convoked
// this spell connives" (#3172).

// Two creatures convoke: each connives in turn, completing its own
// draw-discard-counter before the next draws. The first discards a
// nonland card (counter); the second a land (no counter). A creature
// that did not convoke does nothing.
func TestLethalSchemeConvokersConniveInTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	c1 := pushVanillaCreature(g, me.ID, "Convoker One", 1, 1)
	c2 := pushVanillaCreature(g, me.ID, "Convoker Two", 1, 1)
	bystander := pushVanillaCreature(g, me.ID, "Bystander", 1, 1)
	victim := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	nonland := handCardFull(me, "Spell A", "Sorcery", "{1}", "", nil)
	land := handCardFull(me, "Land B", "Basic Land — Forest", "", "", nil)

	toMainForCost(t, g)
	scheme := handCardFull(me, "Lethal Scheme", "Instant", "{2}{B}{B}", lethalSchemeOracle, []string{"B"})
	floatForTest(g, me, "BB")
	if err := g.CastSpell(me.ID, scheme, game.CastSpellParams{
		Strict: true, TapIDs: []uuid.UUID{c1, c2}, Targets: cardTarget(victim),
	}); err != nil {
		t.Fatalf("cast with two convokers: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefieldByID(g, victim) {
		t.Fatal("the target should be destroyed first")
	}

	// First convoker: one draw, then its discard prompt.
	handBefore := me.Hand.Size()
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("the first convoker owes a discard")
	}
	if handBefore != 3 {
		t.Fatalf("hand while the first prompt is open: %d, want 3 (two seeded + one draw; the second convoker has not drawn yet)", handBefore)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{nonland}); err != nil {
		t.Fatalf("discard the nonland card: %v", err)
	}
	if got := counterCount(g, c1, game.CounterPlusOne); got != 1 {
		t.Errorf("first convoker counters: %d, want 1 (it discarded a nonland card)", got)
	}

	// Second convoker.
	pick = latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("the second convoker owes a discard")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("discard the land: %v", err)
	}
	if got := counterCount(g, c2, game.CounterPlusOne); got != 0 {
		t.Errorf("second convoker counters: %d, want 0 (it discarded a land)", got)
	}
	if got := counterCount(g, bystander, game.CounterPlusOne); got != 0 {
		t.Errorf("a creature that did not convoke connived: %d counters", got)
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) != nil {
		t.Error("no third prompt: only two creatures convoked")
	}
}

// Cast without convoking, nothing connives.
func TestLethalSchemeWithoutConvokeDoesNotConnive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	c1 := pushVanillaCreature(g, me.ID, "Idle", 1, 1)
	victim := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	toMainForCost(t, g)
	scheme := handCardFull(me, "Lethal Scheme", "Instant", "{2}{B}{B}", lethalSchemeOracle, []string{"B"})
	hand := me.Hand.Size()
	floatForTest(g, me, "CCBB")
	if err := g.CastSpell(me.ID, scheme, game.CastSpellParams{Strict: true, Targets: cardTarget(victim)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) != nil {
		t.Error("no creature convoked, so nothing connives")
	}
	if me.Hand.Size() != hand-1 {
		t.Errorf("hand %d, want %d: no draw happened", me.Hand.Size(), hand-1)
	}
	if counterCount(g, c1, game.CounterPlusOne) != 0 {
		t.Error("an idle creature got a counter")
	}
}
