package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestTarmogoyfCDAEmptyGraveyards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[0].ID
	goyf := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Tarmogoyf",
		TypeLine:   "Creature — Lhurgoyf",
		OracleID:   effectstest.TarmogoyfOracle,
		Owner:      owner,
		Controller: owner,
	})

	if got := effectstest.EffectivePower(t, g, goyf); got != 0 {
		t.Errorf("Tarmogoyf power empty graveyards = %d, want 0", got)
	}
	if got := effectstest.EffectiveToughness(t, g, goyf); got != 1 {
		t.Errorf("Tarmogoyf toughness empty graveyards = %d, want 1 (0 + 1)", got)
	}
}

func TestTarmogoyfCDAFourTypesInGraveyards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[0].ID

	// Seed graveyards: Creature, Instant, Sorcery, Land — 4 types.
	effectstest.PushGraveyardCardWithTypeLine(g, owner, "Grizzly Bears", "Creature — Bear")
	effectstest.PushGraveyardCardWithTypeLine(g, owner, "Lightning Bolt", "Instant")
	effectstest.PushGraveyardCardWithTypeLine(g, g.Seats[1].ID, "Wrath of God", "Sorcery") // opponent's graveyard
	effectstest.PushGraveyardCardWithTypeLine(g, g.Seats[1].ID, "Forest", "Basic Land — Forest")

	goyf := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Tarmogoyf",
		TypeLine:   "Creature — Lhurgoyf",
		OracleID:   effectstest.TarmogoyfOracle,
		Owner:      owner,
		Controller: owner,
	})

	if got := effectstest.EffectivePower(t, g, goyf); got != 4 {
		t.Errorf("Tarmogoyf power = %d, want 4 (Creature + Instant + Sorcery + Land)", got)
	}
	if got := effectstest.EffectiveToughness(t, g, goyf); got != 5 {
		t.Errorf("Tarmogoyf toughness = %d, want 5", got)
	}
}

func TestTarmogoyfCDAUpdatesAfterMill(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[0].ID
	effectstest.PushGraveyardCardWithTypeLine(g, owner, "Grizzly Bears", "Creature — Bear")
	effectstest.PushGraveyardCardWithTypeLine(g, owner, "Lightning Bolt", "Instant")
	goyf := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Tarmogoyf",
		TypeLine:   "Creature — Lhurgoyf",
		OracleID:   effectstest.TarmogoyfOracle,
		Owner:      owner,
		Controller: owner,
	})

	if got := effectstest.EffectivePower(t, g, goyf); got != 2 {
		t.Fatalf("Tarmogoyf power pre-mill = %d, want 2 (sanity)", got)
	}

	// Mill an enchantment into a graveyard — should bump goyf to 3/4.
	effectstest.PushGraveyardCardWithTypeLine(g, owner, "Pacifism", "Enchantment — Aura")

	if got := effectstest.EffectivePower(t, g, goyf); got != 3 {
		t.Errorf("Tarmogoyf power post-mill = %d, want 3 (added Enchantment type)", got)
	}
	if got := effectstest.EffectiveToughness(t, g, goyf); got != 4 {
		t.Errorf("Tarmogoyf toughness post-mill = %d, want 4", got)
	}
}

func TestTarmogoyfCDAUnionsAcrossPlayers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[0].ID

	// Seat 0 has a Creature; seat 1 has an Instant; seat 2 has the
	// SAME Creature type as seat 0 (should not double-count).
	effectstest.PushGraveyardCardWithTypeLine(g, owner, "Grizzly Bears", "Creature — Bear")
	effectstest.PushGraveyardCardWithTypeLine(g, g.Seats[1].ID, "Lightning Bolt", "Instant")
	effectstest.PushGraveyardCardWithTypeLine(g, g.Seats[2].ID, "Llanowar Elves", "Creature — Elf Druid")

	goyf := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Tarmogoyf",
		TypeLine:   "Creature — Lhurgoyf",
		OracleID:   effectstest.TarmogoyfOracle,
		Owner:      owner,
		Controller: owner,
	})

	if got := effectstest.EffectivePower(t, g, goyf); got != 2 {
		t.Errorf("Tarmogoyf power = %d, want 2 (Creature + Instant; second Creature doesn't double-count)", got)
	}
}

// TestCommanderIdentityRegressionBant verifies the layer-aware
// commanderIdentityFor still returns the commander's color
// identity correctly for a Bant ({W}{U}{G}) commander — exercises
// the S15 hand-off documented in the sprint plan.
func TestCommanderIdentityRegressionBant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[0]
	// Replace the demo-seed commander wholesale so the identity
	// computation reads the Bant cost we want to test against.
	owner.Command.Cards = nil
	owner.Command.PushTop(game.Card{
		InstanceID:  uuid.New(),
		Name:        "Test Bant Commander",
		TypeLine:    "Legendary Creature — Test",
		ManaCost:    "{1}{G}{W}{U}",
		IsCommander: true,
		Owner:       owner.ID,
		Controller:  owner.ID,
	})

	identity := game.CommanderIdentityForTest(g, owner)
	want := map[string]bool{"W": true, "U": true, "G": true}
	got := map[string]bool{}
	for _, c := range identity {
		got[c] = true
	}
	for c := range want {
		if !got[c] {
			t.Errorf("identity missing %q (got %v, want %v)", c, identity, want)
		}
	}
	for c := range got {
		if !want[c] {
			t.Errorf("identity has unexpected %q (got %v, want %v)", c, identity, want)
		}
	}
}
