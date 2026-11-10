package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const wickOracle = "5af729e8-8a40-4991-8270-612e377451de"

func TestWickIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(wickOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Wick: %+v %v", spec, ok)
	}
}

// wickCastRat casts a vanilla Rat from hand and settles the stack.
func wickCastRat(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Rat", TypeLine: "Creature — Rat", ManaCost: "{B}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	})
	floatForTest(g, me, "B")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
}

func TestWickMakesASnailWhenYouControlNone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dskB1Creature(g, me.ID, "Wick, the Whorled Mind", "Legendary Creature — Rat Warlock", wickOracle, 2, 4)
	advanceToMain(t, g)
	wickCastRat(t, g, me)
	if n := countTokensControlled(g, me.ID, "Snail"); n != 1 {
		t.Fatalf("Snails: %d, want 1", n)
	}
}

func TestWickPutsACounterOnTheSnailWhenYouControlOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dskB1Creature(g, me.ID, "Wick, the Whorled Mind", "Legendary Creature — Rat Warlock", wickOracle, 2, 4)
	snail := b12Creature(g, me.ID, "Snail", "Creature — Snail", 1, 1)
	advanceToMain(t, g)
	wickCastRat(t, g, me)
	if n := countTokensControlled(g, me.ID, "Snail"); n != 0 {
		t.Errorf("a second Snail token was made")
	}
	if got := counterOn(g, snail, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters on the Snail: %d, want 1", got)
	}
}

func TestWickAsksWhichSnailGetsTheCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dskB1Creature(g, me.ID, "Wick, the Whorled Mind", "Legendary Creature — Rat Warlock", wickOracle, 2, 4)
	a := b12Creature(g, me.ID, "Snail A", "Creature — Snail", 1, 1)
	b := b12Creature(g, me.ID, "Snail B", "Creature — Snail", 1, 1)
	advanceToMain(t, g)
	wickCastRat(t, g, me)
	answerChooseCards(t, g, me.ID, b)
	if counterOn(g, b, game.CounterPlusOne) != 1 || counterOn(g, a, game.CounterPlusOne) != 0 {
		t.Errorf("counters: A=%d B=%d, want the chosen Snail B only", counterOn(g, a, game.CounterPlusOne), counterOn(g, b, game.CounterPlusOne))
	}
}

func TestWickSacrificeASnailDamagesEachOpponentAndDrawsItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wick := dskB1Creature(g, me.ID, "Wick, the Whorled Mind", "Legendary Creature — Rat Warlock", wickOracle, 2, 4)
	snail := b12Creature(g, me.ID, "Snail", "Creature — Snail", 1, 1)
	g.WithWriteLock(func() { _ = g.AddCounterByKeyedForEffect(uuid.Nil, snail, game.CounterPlusOne, 2, game.KeyedTail{}) })
	seedLibrary(me, "a", "b", "c", "d", "e")
	advanceToMain(t, g)
	floatForTest(g, me, "UBR")
	hand, life := me.Hand.Size(), opp.Life
	if err := g.ActivateCatalogAbility(me.ID, wick, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{snail}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("opponent lost %d, want 3 (the Snail's power as it last existed)", got)
	}
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
}

func TestWickCantSacrificeANonSnail(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	wick := dskB1Creature(g, me.ID, "Wick, the Whorled Mind", "Legendary Creature — Rat Warlock", wickOracle, 2, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	floatForTest(g, me, "UBR")
	if err := g.ActivateCatalogAbility(me.ID, wick, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Error("sacrificed a Bear to a Snail cost")
	}
}
