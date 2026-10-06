package effects

import (
	"testing"

	"github.com/google/uuid"

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

// windowAt reports whether the exiled card's grant is live for `me`.
func windowAt(g *game.Game, id uuid.UUID, me uuid.UUID) bool {
	perm := g.CastPermissionOnCardByIDForEffect(id)
	return perm != nil && permissionLive(g, perm, me)
}

func TestObNixilisKingpinWindowClosesAtThisTurnsEndStep(t *testing.T) {
	g, me, ob, top := obKingpinBoard(t)
	g.WithWriteLock(func() {
		for _, opp := range g.Seats[1:3] {
			_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1)
		}
	})
	settleLifeBoundary(g)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) || countersOn(g, ob, "+1/+1") != 1 {
		t.Fatal("trigger did not resolve")
	}
	if !windowAt(g, top, me.ID) {
		t.Fatal("the window is closed on the turn it was made")
	}
	advanceToStepOf(t, g, nextEndStepSeatNumber, game.StepEnd)
	if windowAt(g, top, me.ID) {
		t.Error("the card is still playable in the end step the window closes at")
	}
}

func TestHasteMagicPumpsExilesAndWindowEndsAtTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	top := stackLibrary(me, "Top Card", "Sorcery", "{1}")

	castCatalogSpell(t, g, "Haste Magic", "Instant", hasteMagicOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 5 || tough != 3 {
		t.Errorf("P/T = %d/%d, want 5/3", p, tough)
	}
	if !effectiveAbilitiesContain(t, g, bear, "haste") {
		t.Error("target did not gain haste")
	}
	if !g.Exile.Contains(top) || !windowAt(g, top, me.ID) {
		t.Fatal("the top card was not exiled with a live play window")
	}
	advanceToStepOf(t, g, g.Turn.ActiveSeat, game.StepEnd)
	if windowAt(g, top, me.ID) {
		t.Error("the window outlived the end step")
	}
}

func TestHasteMagicOnAnOpponentsTurnLastsToYourNextEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToStepOf(t, g, 1, game.StepPrecombatMain)
	opp := g.Seats[1]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	top := stackLibrary(me, "Top Card", "Sorcery", "{1}")
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
	passPriorityAroundTable(t, g)

	if !windowAt(g, top, me.ID) {
		t.Fatal("no live window after the cast")
	}
	advanceToStepOf(t, g, 1, game.StepEnd)
	if !windowAt(g, top, me.ID) {
		t.Fatal("the opponent's end step closed my window")
	}
	advanceToStepOf(t, g, 0, game.StepEnd)
	if windowAt(g, top, me.ID) {
		t.Error("the window outlived my next end step")
	}
}

func TestWiccanExilesOnNoncreatureCastOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Wiccan, Young Avenger", "Legendary Creature — Mutant Warlock Hero", wiccanOracle, false)
	top := stackLibrary(me, "Top Card", "Sorcery", "{1}")

	// A creature spell does nothing.
	castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(top) {
		t.Fatal("a creature spell triggered Wiccan")
	}

	// A noncreature spell does.
	castCatalogSpell(t, g, "Shock-ish", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) || !windowAt(g, top, me.ID) {
		t.Fatal("a noncreature spell did not exile the top card with a play window")
	}
	advanceToStepOf(t, g, g.Turn.ActiveSeat, game.StepEnd)
	if windowAt(g, top, me.ID) {
		t.Error("the window outlived the end step")
	}
}

func TestOperaLoveSongExileModeAndPumpMode(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	top := stackLibrary(me, "Second", "Sorcery", "{1}")
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "First", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	first := me.Library.Cards[0].InstanceID

	castModal(t, g, "Opera Love Song", "Instant", operaLoveSongOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{first, top} {
		if !g.Exile.Contains(id) || !windowAt(g, id, me.ID) {
			t.Fatalf("card %s not exiled with a live window", id)
		}
	}
	advanceToStepOf(t, g, g.Turn.ActiveSeat, game.StepEnd)
	for _, id := range []uuid.UUID{first, top} {
		if windowAt(g, id, me.ID) {
			t.Error("a window outlived the end step")
		}
	}
}

func TestOperaLoveSongPumpModeTwoTargets(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "A", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	b := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "B", TypeLine: "Creature — Bear",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	top := stackLibrary(me, "Top", "Sorcery", "{1}")

	castModal(t, g, "Opera Love Song", "Instant", operaLoveSongOracle, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	passPriorityAroundTable(t, g)
	if effectivePower(t, g, a) != 4 || effectivePower(t, g, b) != 3 {
		t.Errorf("powers = %d, %d; want 4, 3", effectivePower(t, g, a), effectivePower(t, g, b))
	}
	if g.Exile.Contains(top) {
		t.Error("the pump mode exiled a card")
	}
}
