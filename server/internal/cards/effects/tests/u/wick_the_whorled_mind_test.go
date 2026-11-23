package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const wickOracle = "5af729e8-8a40-4991-8270-612e377451de"

func TestWickIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(wickOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Wick: %+v %v", spec, ok)
	}
}

func TestWickMakesASnailWhenYouControlNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.DskB1Creature(g, me.ID, "Wick, the Whorled Mind", "Legendary Creature — Rat Warlock", wickOracle, 2, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.WickCastRat(t, g, me)
	if n := effectstest.CountTokensControlled(g, me.ID, "Snail"); n != 1 {
		t.Fatalf("Snails: %d, want 1", n)
	}
}

func TestWickPutsACounterOnTheSnailWhenYouControlOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.DskB1Creature(g, me.ID, "Wick, the Whorled Mind", "Legendary Creature — Rat Warlock", wickOracle, 2, 4)
	snail := effectstest.B12Creature(g, me.ID, "Snail", "Creature — Snail", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.WickCastRat(t, g, me)
	if n := effectstest.CountTokensControlled(g, me.ID, "Snail"); n != 0 {
		t.Errorf("a second Snail token was made")
	}
	if got := effectstest.CounterOn(g, snail, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters on the Snail: %d, want 1", got)
	}
}

func TestWickAsksWhichSnailGetsTheCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.DskB1Creature(g, me.ID, "Wick, the Whorled Mind", "Legendary Creature — Rat Warlock", wickOracle, 2, 4)
	a := effectstest.B12Creature(g, me.ID, "Snail A", "Creature — Snail", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "Snail B", "Creature — Snail", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.WickCastRat(t, g, me)
	effectstest.AnswerChooseCards(t, g, me.ID, b)
	if effectstest.CounterOn(g, b, game.CounterPlusOne) != 1 || effectstest.CounterOn(g, a, game.CounterPlusOne) != 0 {
		t.Errorf("counters: A=%d B=%d, want the chosen Snail B only", effectstest.CounterOn(g, a, game.CounterPlusOne), effectstest.CounterOn(g, b, game.CounterPlusOne))
	}
}

func TestWickSacrificeASnailDamagesEachOpponentAndDrawsItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wick := effectstest.DskB1Creature(g, me.ID, "Wick, the Whorled Mind", "Legendary Creature — Rat Warlock", wickOracle, 2, 4)
	snail := effectstest.B12Creature(g, me.ID, "Snail", "Creature — Snail", 1, 1)
	g.WithWriteLock(func() { _ = g.AddCounterByKeyedForEffect(uuid.Nil, snail, game.CounterPlusOne, 2, game.KeyedTail{}) })
	effectstest.SeedLibrary(me, "a", "b", "c", "d", "e")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "UBR")
	hand, life := me.Hand.Size(), opp.Life
	if err := g.ActivateCatalogAbility(me.ID, wick, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{snail}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("opponent lost %d, want 3 (the Snail's power as it last existed)", got)
	}
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
}

func TestWickCantSacrificeANonSnail(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wick := effectstest.DskB1Creature(g, me.ID, "Wick, the Whorled Mind", "Legendary Creature — Rat Warlock", wickOracle, 2, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "UBR")
	if err := g.ActivateCatalogAbility(me.ID, wick, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Error("sacrificed a Bear to a Snail cost")
	}
}
