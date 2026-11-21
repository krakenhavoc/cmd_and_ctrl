package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #2373 — "until your next end step" cards: Ob Nixilis, Captive
// Kingpin, Haste Magic, Wiccan, Young Avenger, Opera Love Song. The
// duration itself is tested in game/duration_next_end_step_test.go;
// these pin that each card stamps it and that the window closes as the
// end step BEGINS.

const (
	hasteMagicOracle      = "5249dd0d-3bda-452e-ac6e-774c0ebc2e41"
	wiccanOracle          = "a317d5f4-a998-4f72-b080-5345bc7cc668"
	operaLoveSongOracle   = "22f259b3-9de1-44a1-a89d-f2f1b1249d4d"
	nextEndStepSeatNumber = 0
)

func TestObNixilisKingpinWindowClosesAtThisTurnsEndStep(t *testing.T) {
	g, me, ob, top := effectstest.ObKingpinBoard(t)
	g.WithWriteLock(func() {
		for _, opp := range g.Seats[1:3] {
			_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1)
		}
	})
	effectstest.SettleLifeBoundary(g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) || effectstest.CountersOn(g, ob, "+1/+1") != 1 {
		t.Fatal("trigger did not resolve")
	}
	if !effectstest.WindowAt(g, top, me.ID) {
		t.Fatal("the window is closed on the turn it was made")
	}
	effectstest.AdvanceToStepOf(t, g, nextEndStepSeatNumber, game.StepEnd)
	if effectstest.WindowAt(g, top, me.ID) {
		t.Error("the card is still playable in the end step the window closes at")
	}
}

func TestHasteMagicPumpsExilesAndWindowEndsAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	top := effectstest.StackLibrary(me, "Top Card", "Sorcery", "{1}")

	effectstest.CastCatalogSpell(t, g, "Haste Magic", "Instant", hasteMagicOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 5 || tough != 3 {
		t.Errorf("P/T = %d/%d, want 5/3", p, tough)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "haste") {
		t.Error("target did not gain haste")
	}
	if !g.Exile.Contains(top) || !effectstest.WindowAt(g, top, me.ID) {
		t.Fatal("the top card was not exiled with a live play window")
	}
	effectstest.AdvanceToStepOf(t, g, g.Turn.ActiveSeat, game.StepEnd)
	if effectstest.WindowAt(g, top, me.ID) {
		t.Error("the window outlived the end step")
	}
}

func TestHasteMagicOnAnOpponentsTurnLastsToYourNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToStepOf(t, g, 1, game.StepPrecombatMain)
	opp := g.Seats[1]
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	top := effectstest.StackLibrary(me, "Top Card", "Sorcery", "{1}")
	// Something to draw on my next turn once the top card is exiled.
	me.Library.PushBottom(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})

	// Seat 0 casts on seat 1's turn: seed it into seat 0's hand and cast
	// at instant speed.
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Haste Magic", TypeLine: "Instant",
		OracleID: hasteMagicOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.WindowAt(g, top, me.ID) {
		t.Fatal("no live window after the cast")
	}
	effectstest.AdvanceToStepOf(t, g, 1, game.StepEnd)
	if !effectstest.WindowAt(g, top, me.ID) {
		t.Fatal("the opponent's end step closed my window")
	}
	effectstest.AdvanceToStepOf(t, g, 0, game.StepEnd)
	if effectstest.WindowAt(g, top, me.ID) {
		t.Error("the window outlived my next end step")
	}
}

func TestWiccanExilesOnNoncreatureCastOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Wiccan, Young Avenger", "Legendary Creature — Mutant Warlock Hero", wiccanOracle, false)
	top := effectstest.StackLibrary(me, "Top Card", "Sorcery", "{1}")

	// A creature spell does nothing.
	effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(top) {
		t.Fatal("a creature spell triggered Wiccan")
	}

	// A noncreature spell does.
	effectstest.CastCatalogSpell(t, g, "Shock-ish", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) || !effectstest.WindowAt(g, top, me.ID) {
		t.Fatal("a noncreature spell did not exile the top card with a play window")
	}
	effectstest.AdvanceToStepOf(t, g, g.Turn.ActiveSeat, game.StepEnd)
	if effectstest.WindowAt(g, top, me.ID) {
		t.Error("the window outlived the end step")
	}
}

func TestOperaLoveSongExileModeAndPumpMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	top := effectstest.StackLibrary(me, "Second", "Sorcery", "{1}")
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "First", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	first := me.Library.Cards[0].InstanceID

	effectstest.CastModal(t, g, "Opera Love Song", "Instant", operaLoveSongOracle, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{first, top} {
		if !g.Exile.Contains(id) || !effectstest.WindowAt(g, id, me.ID) {
			t.Fatalf("card %s not exiled with a live window", id)
		}
	}
	effectstest.AdvanceToStepOf(t, g, g.Turn.ActiveSeat, game.StepEnd)
	for _, id := range []uuid.UUID{first, top} {
		if effectstest.WindowAt(g, id, me.ID) {
			t.Error("a window outlived the end step")
		}
	}
}

func TestOperaLoveSongPumpModeTwoTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "A", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	b := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "B", TypeLine: "Creature — Bear",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	top := effectstest.StackLibrary(me, "Top", "Sorcery", "{1}")

	effectstest.CastModal(t, g, "Opera Love Song", "Instant", operaLoveSongOracle, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EffectivePower(t, g, a) != 4 || effectstest.EffectivePower(t, g, b) != 3 {
		t.Errorf("powers = %d, %d; want 4, 3", effectstest.EffectivePower(t, g, a), effectstest.EffectivePower(t, g, b))
	}
	if g.Exile.Contains(top) {
		t.Error("the pump mode exiled a card")
	}
}
