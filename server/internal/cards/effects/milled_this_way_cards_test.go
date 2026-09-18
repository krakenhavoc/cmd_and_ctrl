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

// TestOonaMakesNoFaerieForACommanderThatTookTheCommandZone is the
// issue's own card. Oona exiles two blue cards off an opponent's
// library, one of them their commander: that commander is asked about
// the command zone, the Faeries wait for the answer, and taking the
// offer is one fewer Faerie because the card never reached exile.
func TestOonaMakesNoFaerieForACommanderThatTookTheCommandZone(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
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

	if n := countFaeries(g, me.ID); n != 0 {
		t.Fatalf("%d Faeries while the CR 903.9 prompt is open, want none — "+
			"what was exiled this way is not known yet", n)
	}
	b36AcceptCommandZone(t, g, them.ID)

	if !them.Command.Contains(commander) {
		t.Fatal("the commander took the offer and is in the command zone")
	}
	if n := countFaeries(g, me.ID); n != 1 {
		t.Errorf("Faerie Rogues = %d, want 1 — the commander went to the command zone, not to exile, "+
			"so it was not exiled this way (CR 400.7)", n)
	}
}

// TestOonaMakesAFaerieForACommanderThatWentToExile is the other
// answer: declining really does exile the card, so it was exiled this
// way and it pays for its Faerie — an action later than it used to.
func TestOonaMakesAFaerieForACommanderThatWentToExile(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
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
	b21DeclineCommandZone(t, g, them.ID)

	if !g.Exile.Contains(commander) {
		t.Fatal("declining exiles the commander after all")
	}
	if n := countFaeries(g, me.ID); n != 2 {
		t.Errorf("Faerie Rogues = %d, want 2 — both blue cards reached exile", n)
	}
}

// TestHelmOfObedienceWaitsForAMilledCommandersAnswer — the Helm reads
// "a creature card put into their graveyard", which a commander only
// is once its owner declines the command zone. The old read-back
// happened with that prompt still open, found nothing, and dropped the
// Helm's own second half on the floor.
func TestHelmOfObedienceWaitsForAMilledCommandersAnswer(t *testing.T) {
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

	if !g.Battlefield.Contains(helm) {
		t.Fatal("the Helm is not sacrificed while the CR 903.9 prompt is open — nothing has been milled yet")
	}
	b21DeclineCommandZone(t, g, opp.ID)

	if !g.Battlefield.Contains(commander) {
		t.Fatalf("declining puts the commander in the graveyard, so the Helm reanimates it")
	}
	if c := findTestCard(g, commander); c == nil || c.Controller != me.ID {
		t.Error(`the reanimated creature arrives under the activator's control ("under your control")`)
	}
	if g.Battlefield.Contains(helm) {
		t.Error("finding a creature sacrifices the Helm")
	}
}
