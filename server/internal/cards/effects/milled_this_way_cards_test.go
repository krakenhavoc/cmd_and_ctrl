package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// milled_this_way_cards_test.go — #893 at the catalog level.
//
// The engine half (game/milled_this_way_test.go) taught the mill to
// report what LANDED in the zone it aimed at; this is the same rule on
// the cards that read the mill's answer. It is not the answer they used
// to get: the old list held every leg that had not paused, so a card a
// graveyard replacement sent to exile, and a commander whose owner was
// still being ASKED about the command zone, both counted.
//
// CR 400.7: the object that ARRIVED is the one the effect moved.
//
// ADR 0115: a commander no longer pauses a mill or an exile. It arrives
// like any other card, so it counts, and CR 903.9a asks its owner about
// the command zone afterwards.

// libraryCardFor pushes a card onto the top of p's library. The top is
// the LAST element, so the last push is the first card milled.
func libraryCardFor(p *game.Player, name, typeLine string, colors ...string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, Colors: colors,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

// countFaeries counts the Faerie Rogue tokens `controller` controls.
func countFaeries(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Faerie Rogue" && c.Controller == controller {
			n++
		}
	}
	return n
}

// oonaExilesTheirCommander activates Oona at X=2 against an opponent
// whose top two cards are blue, one of them their commander, and
// resolves it. Returns the commander's ID.
func oonaExilesTheirCommander(t *testing.T, g *game.Game, me, them *game.Player) uuid.UUID {
	t.Helper()
	aangAdvanceToMain(t, g, 0)
	oona := b31Push(g, me.ID, "Oona, Queen of the Fae", "Legendary Creature — Faerie Wizard",
		"6052822d-47a2-4d69-a32d-40cdd600d7a9", "{3}{U/B}{U/B}{U/B}", 5, 5, "U", "B")
	libraryCardFor(them, "Brainstorm", "Instant", "U")
	commander := libraryCardFor(them, "Their Commander", "Legendary Creature — Merfolk", "U")
	markCommanderCard(t, g, them, commander)
	b31AddMana(me, "U", "U", "U")

	if err := g.ActivateCatalogAbility(me.ID, oona, 0, game.ActivateAbilityParams{
		XValue:  2,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerColor(t, g, me.ID, "U")
	return commander
}

// TestOonaMakesAFaerieForACommanderThatThenGoesHome is the issue's own
// card. Oona exiles two blue cards off an opponent's library, one of
// them their commander. Since ADR 0115 the commander is exiled like the
// other card, so both pay for a Faerie; its owner is then asked about
// the command zone (CR 903.9a), and taking it does not undo the exile.
func TestOonaMakesAFaerieForACommanderThatThenGoesHome(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	commander := oonaExilesTheirCommander(t, g, me, them)

	if n := countFaeries(g, me.ID); n != 2 {
		t.Fatalf("Faerie Rogues = %d before the CR 903.9a answer, want 2 — both blue cards reached exile", n)
	}
	answerCommanderReturn(t, g, them.ID, true)
	if !them.Command.Contains(commander) {
		t.Fatal("the commander took the offer and is in the command zone")
	}
	if n := countFaeries(g, me.ID); n != 2 {
		t.Errorf("Faerie Rogues = %d after the answer, want 2", n)
	}
}

// TestOonaMakesAFaerieForACommanderThatWentToExile is the other
// answer: declined, the commander stays in exile, and it paid for its
// Faerie when it got there.
func TestOonaMakesAFaerieForACommanderThatWentToExile(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	commander := oonaExilesTheirCommander(t, g, me, them)
	answerCommanderReturn(t, g, them.ID, false)

	if !g.Exile.Contains(commander) {
		t.Fatal("declining leaves the commander in exile")
	}
	if n := countFaeries(g, me.ID); n != 2 {
		t.Errorf("Faerie Rogues = %d, want 2 — both blue cards reached exile", n)
	}
}

// TestHelmOfObedienceReanimatesAMilledCommander — the Helm reads "a
// creature card put into their graveyard". Before ADR 0115 a commander
// was one only once its owner declined the command zone; now it is
// milled into the graveyard like any other card, so the Helm stops on
// it, is sacrificed, and puts it onto the battlefield under its
// controller. It left the graveyard within the resolution, so CR 903.9a
// never asks about it (CR 704.4).
func TestHelmOfObedienceReanimatesAMilledCommander(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fillPool(me, 5)
	helm := pushCatalogPermanent(g, me.ID, "Helm of Obedience", "Artifact", helmOfObedienceOracle, false)

	commander := libraryCardFor(opp, "Their Commander", "Legendary Creature — Beast")
	markCommanderCard(t, g, opp, commander)
	libraryCardFor(opp, "Filler", "Sorcery")

	if err := g.ActivateCatalogAbility(me.ID, helm, 0, game.ActivateAbilityParams{
		XValue:  5,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		Strict:  true,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(commander) {
		t.Fatalf("the milled commander is in %s, want the battlefield", b12ZoneOf(g, commander))
	}
	if c := findTestCard(g, commander); c == nil || c.Controller != me.ID {
		t.Error(`the reanimated creature arrives under the activator's control ("under your control")`)
	}
	if g.Battlefield.Contains(helm) {
		t.Error("finding a creature sacrifices the Helm")
	}
	if commanderReturnPromptFor(g, opp.ID) != nil {
		t.Error("a commander that left the graveyard within the resolution was offered the command zone")
	}
}
