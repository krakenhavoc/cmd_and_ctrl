package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const sonicScrewdriverOracle = "cd68cc31-12fd-48ff-b37b-bccfe4172974"

func pushSonicScrewdriver(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sonic Screwdriver", TypeLine: "Artifact",
		OracleID: sonicScrewdriverOracle, Owner: owner, Controller: owner,
	})
}

// TestSonicScrewdriverUntapsAnotherArtifact — the {1},{T} ability.
func TestSonicScrewdriverUntapsAnotherArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	screwdriver := pushSonicScrewdriver(g, me.ID)
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID, Tapped: true,
	})

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, screwdriver, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility(untap): %v", err)
	}
	passPriorityAroundTable(t, g)

	if c, ok := g.LookupCardForEffect(rock); !ok || c.Tapped {
		t.Error("the targeted artifact should be untapped")
	}
}

// TestSonicScrewdriverScries1 — the {2},{T} ability.
func TestSonicScrewdriverScries1(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Bottom Me")
	screwdriver := pushSonicScrewdriver(g, me.ID)

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, screwdriver, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility(scry): %v", err)
	}
	passPriorityAroundTable(t, g)

	if c := scryChoiceFor(g, me.ID); c == nil {
		t.Fatalf("no scry-1 prompt: %+v", g.PendingChoices)
	}
}

// TestSonicScrewdriverMakesACreatureUnblockable — the {3},{T} ability.
func TestSonicScrewdriverMakesACreatureUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	screwdriver := pushSonicScrewdriver(g, me.ID)
	attacker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Attacker", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	_ = opp

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, screwdriver, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: attacker}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility(unblockable): %v", err)
	}
	passPriorityAroundTable(t, g)

	assertRestrictions(t, g, attacker, game.CantBeBlocked)
}

// TestSonicScrewdriverManaAbility — the {T}: any color.
func TestSonicScrewdriverManaAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	screwdriver := pushSonicScrewdriver(g, me.ID)

	if err := g.ActivateManaAbility(me.ID, screwdriver, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	found := false
	for _, tok := range me.ManaPool {
		if tok.Color == "U" {
			found = true
		}
	}
	if !found {
		t.Errorf("mana pool after activating for blue: %+v", me.ManaPool)
	}
}
