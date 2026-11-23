package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// redcapTurnedBlueAndKilled is redcapTurnedBlueThenKilled without the
// final resolution: the trigger is still on the stack, so a test can
// respond to it once more (#1429). Returns the Redcap's instance ID.
func RedcapTurnedBlueAndKilled(t *testing.T, g *game.Game, target game.TargetRef) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	redcap := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: redcap, Name: "Murderous Redcap", TypeLine: "Creature — Goblin Assassin",
		OracleID: B40MurderousRedcapOracle, Power: 2, Toughness: 2, Colors: []string{"B", "R"},
		Owner: me.ID, Controller: me.ID,
	})
	AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, redcap, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Redcap: %v", err)
	}
	PassUntilOnBattlefield(t, g, redcap)
	B04WaitForPick(t, g, me.ID)
	p := LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, target); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	if TriggerOnStack(g, redcap) == nil {
		t.Fatal("setup: the Redcap's enter trigger should be waiting")
	}

	wisps := CastCatalogSpell(t, g, "Cerulean Wisps", "Instant", B43CeruleanWispsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: redcap}})
	for i := 0; i < 8 && g.Stack.Contains(wisps); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if g.Stack.Contains(wisps) || TriggerOnStack(g, redcap) == nil {
		t.Fatal("setup: the Wisps should have resolved with the Redcap trigger still waiting")
	}
	var colors []string
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, _ := g.LookupCardForEffect(redcap)
		colors = c.Effective().Colors
	})
	if len(colors) != 1 || colors[0] != "U" {
		t.Fatalf("setup: the Redcap is %v on the battlefield, want blue", colors)
	}

	RemoveInResponse(t, g, redcap, false)
	var inYard game.Card
	g.WithWriteLock(func() { inYard, _ = g.LookupCardForEffect(redcap) })
	if !inYard.HasColor("R") {
		t.Fatal("setup: the Redcap card in the graveyard must be red, or these tests prove nothing")
	}
	// #2075: the Redcap has persist now, and its persist trigger went on
	// the stack above the enter trigger. These tests are about the enter
	// trigger's departed source, so the card leaves the graveyard (it is
	// a new object, black-red in exile too) and the persist trigger
	// resolves doing nothing, leaving the board as it was before persist.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneExile}, redcap); err != nil {
		t.Fatalf("setup: exile the Redcap from the graveyard: %v", err)
	}
	for i := 0; i < 8 && DiesReturnItems(g) > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("setup: resolve the persist trigger: %v", err)
		}
	}
	if DiesReturnItems(g) > 0 || TriggerOnStack(g, redcap) == nil {
		t.Fatal("setup: the persist trigger should be gone and the enter trigger still waiting")
	}
	return redcap
}
