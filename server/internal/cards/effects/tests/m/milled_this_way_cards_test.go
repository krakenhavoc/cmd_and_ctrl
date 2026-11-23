package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
	effectstest.AangAdvanceToMain(t, g, 0)
	oona := effectstest.B31Push(g, me.ID, "Oona, Queen of the Fae", "Legendary Creature — Faerie Wizard",
		"6052822d-47a2-4d69-a32d-40cdd600d7a9", "{3}{U/B}{U/B}{U/B}", 5, 5, "U", "B")
	effectstest.LibraryCardFor(them, "Brainstorm", "Instant", "U")
	commander := effectstest.LibraryCardFor(them, "Their Commander", "Legendary Creature — Merfolk", "U")
	effectstest.MarkCommanderCard(t, g, them, commander)
	effectstest.B31AddMana(me, "U", "U", "U")

	if err := g.ActivateCatalogAbility(me.ID, oona, 0, game.ActivateAbilityParams{
		XValue:  2,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "U")
	return commander
}

// TestOonaMakesAFaerieForACommanderThatThenGoesHome is the issue's own
// card. Oona exiles two blue cards off an opponent's library, one of
// them their commander. Since ADR 0115 the commander is exiled like the
// other card, so both pay for a Faerie; its owner is then asked about
// the command zone (CR 903.9a), and taking it does not undo the exile.
func TestOonaMakesAFaerieForACommanderThatThenGoesHome(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	commander := oonaExilesTheirCommander(t, g, me, them)

	if n := countFaeries(g, me.ID); n != 2 {
		t.Fatalf("Faerie Rogues = %d before the CR 903.9a answer, want 2 — both blue cards reached exile", n)
	}
	effectstest.AnswerCommanderReturn(t, g, them.ID, true)
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
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	commander := oonaExilesTheirCommander(t, g, me, them)
	effectstest.AnswerCommanderReturn(t, g, them.ID, false)

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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.FillPool(me, 5)
	helm := effectstest.PushCatalogPermanent(g, me.ID, "Helm of Obedience", "Artifact", effectstest.HelmOfObedienceOracle, false)

	commander := effectstest.LibraryCardFor(opp, "Their Commander", "Legendary Creature — Beast")
	effectstest.MarkCommanderCard(t, g, opp, commander)
	effectstest.LibraryCardFor(opp, "Filler", "Sorcery")

	if err := g.ActivateCatalogAbility(me.ID, helm, 0, game.ActivateAbilityParams{
		XValue:  5,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		Strict:  true,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(commander) {
		t.Fatalf("the milled commander is in %s, want the battlefield", effectstest.B12ZoneOf(g, commander))
	}
	if c := effectstest.FindTestCard(g, commander); c == nil || c.Controller != me.ID {
		t.Error(`the reanimated creature arrives under the activator's control ("under your control")`)
	}
	if g.Battlefield.Contains(helm) {
		t.Error("finding a creature sacrifices the Helm")
	}
	if effectstest.CommanderReturnPromptFor(g, opp.ID) != nil {
		t.Error("a commander that left the graveyard within the resolution was offered the command zone")
	}
}
