package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// paused_exile_continuations_test.go — #894. Two single-card exile
// loops that ran on past a leg a commander's command-zone prompt had
// PAUSED (before ADR 0115 made exile and graveyard moves unpaused).
//
// Living Death is the one that lost a card: its exile / sacrifice /
// reanimate swap moved on while a commander card's owner was being
// asked about the command zone, so when they declined, that commander
// arrived in exile with the "put all cards they exiled this way onto
// the battlefield" step already finished — stranded, permanently.
// Path to Exile is the one that only looked wrong: it offered the
// basic-land search while the same question was still on the table.
//
// The shared answer is the continuation the exile already has: the
// batch for Living Death (ExileCardsThenForEffect, which reports the
// cards that really reached exile — CR 400.7), ExileTarget.Then for
// the single-card cases.
//
// ADR 0115 removed the pause these tests were written against: a
// commander is exiled like any other card and CR 903.9a asks its owner
// only at the next state-based action check. The tests now pin the
// cards' printed behaviour on that order. A commander exiled and put
// back within one resolution (Living Death, a flicker) is never asked
// at all (CR 704.4).

// livingDeathBoard seats the swap this file tests: `opp` has a
// commander creature card and a plain creature card in their
// graveyard, and `me` has one creature on the battlefield. Returns the
// commander card and the plain card.
func livingDeathBoard(t *testing.T, g *game.Game, me, opp *game.Player) (commander, dead, alive uuid.UUID) {
	t.Helper()
	commander = b17GraveyardCard(opp, "Their Commander", "Legendary Creature — Human", "{2}{B}")
	markCommanderCard(t, g, opp, commander)
	dead = pushGraveyardCardForTest(opp, "Their Dead")
	alive = seedCreature(g, "My Alive", me.ID)
	return commander, dead, alive
}

// TestLivingDeathReturnsACommanderCardItExiled is the issue's board on
// ADR 0115's order. The commander card is exiled with the rest of the
// graveyard, the living creature is sacrificed, and every card exiled
// this way, the commander included, comes back. It left exile within
// the resolution, so CR 903.9a never asks about it.
func TestLivingDeathReturnsACommanderCardItExiled(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	commander, dead, alive := livingDeathBoard(t, g, me, opp)

	castCatalogSpell(t, g, "Living Death", "Sorcery", b03LivingDeathOracle, nil)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(alive) {
		t.Error("the living creature is sacrificed")
	}
	if !me.Graveyard.Contains(alive) {
		t.Error("the sacrificed creature stays in the graveyard — it was not exiled this way")
	}
	if countBattlefieldNamed(g, opp.ID, "Their Dead") != 1 {
		t.Errorf("the plain graveyard creature comes back; battlefield holds %d",
			countBattlefieldNamed(g, opp.ID, "Their Dead"))
	}
	if g.Exile.Contains(dead) || g.Exile.Contains(commander) {
		t.Error("nothing that was exiled this way is left in exile")
	}
	if countBattlefieldNamed(g, opp.ID, "Their Commander") != 1 {
		t.Error("the commander card was exiled this way, so the swap puts it onto the battlefield")
	}
	if commanderReturnPromptFor(g, opp.ID) != nil {
		t.Error("a commander that left exile within the resolution was offered the command zone")
	}
}

// pathACommander casts Path to Exile on `victim`'s commander and
// resolves it. Returns the commander and the Forest in their library.
func pathACommander(t *testing.T, g *game.Game, victim *game.Player) (commander, forest uuid.UUID) {
	t.Helper()
	commander = b36Commander(g, victim.ID, "Their Commander")
	forest = pushLibraryCardForTest(victim, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	castCatalogSpell(t, g, "Path to Exile", "Instant",
		"d683d985-9888-4d21-8b5f-69e69ce4a03b",
		[]game.TargetRef{{Kind: game.TargetCard, ID: commander}},
	)
	passPriorityAroundTable(t, g)
	return commander, forest
}

// TestPathToExileExilesACommanderThenOffersTheSearch — the ordering
// half of #894 on ADR 0115's order. The commander is exiled, the search
// is offered as part of the same resolution, and CR 903.9a asks its
// owner about the command zone only after that.
func TestPathToExileExilesACommanderThenOffersTheSearch(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[1]
	commander, forest := pathACommander(t, g, victim)

	if !g.Exile.Contains(commander) {
		t.Fatal("Path to Exile exiles the commander")
	}
	if searchChoiceFor(g, victim.ID) == nil {
		t.Fatal("the search is offered once the exile has landed")
	}
	answerSearchByID(t, g, victim.ID, forest)
	if !g.Battlefield.Contains(forest) {
		t.Error("and it still fetches the basic land")
	}
	answerCommanderReturn(t, g, victim.ID, false)
	if !g.Exile.Contains(commander) {
		t.Error("declined, the commander stays in exile")
	}
}

// TestPathToExileStillSearchesWhenTheCommanderThenGoesHome — the search
// is a separate sentence, not an "if you do", and the commander was
// exiled anyway; its owner's later "yes" changes nothing about it.
func TestPathToExileStillSearchesWhenTheCommanderThenGoesHome(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[1]
	commander, forest := pathACommander(t, g, victim)

	answerSearchByID(t, g, victim.ID, forest)
	answerCommanderReturn(t, g, victim.ID, true)
	if !victim.Command.Contains(commander) {
		t.Fatal("the commander took the offer")
	}
	if !g.Battlefield.Contains(forest) {
		t.Error("the search still fetched the basic land")
	}
}

// TestFlickerReturnsACommanderWithNoCommandZoneQuestion — the Flicker
// primitive (Y'shtola Rhul) is Living Death's shape with one card:
// exile it, then put that same card back. The return used to run with
// the CR 903.9 prompt open. Since ADR 0115 the exile lands at once and
// the card is back before any state-based action check, so CR 903.9a
// never sees it (CR 704.4): the commander simply blinks.
func TestFlickerReturnsACommanderWithNoCommandZoneQuestion(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	commander := b36Commander(g, me.ID, "My Commander")

	g.WithWriteLock(func() {
		item := &game.StackItem{Controller: me.ID, SourceCardID: uuid.New()}
		if err := (Flicker{Target: commander}).Apply(ctxFor(g, item)); err != nil {
			t.Fatalf("Flicker: %v", err)
		}
	})

	if g.Battlefield.Contains(commander) {
		t.Error("the flicker really happened: the pre-exile object is gone (CR 400.7)")
	}
	if got := countBattlefieldNamed(g, me.ID, "My Commander"); got != 1 {
		t.Errorf("battlefield holds %d copies of the commander, want 1", got)
	}
	if g.Exile.Contains(commander) {
		t.Error("the flickered commander is not left in exile")
	}
	if commanderReturnPromptFor(g, me.ID) != nil {
		t.Error("a commander exiled and returned within one resolution was offered the command zone")
	}
}
