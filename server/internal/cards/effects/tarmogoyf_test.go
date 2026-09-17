package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tarmogoyf_test.go covers the S16 sub-PR 5 Tarmogoyf catalog entry
// (Layer 7a CDA): power = distinct card types in all graveyards,
// toughness = power + 1.

const tarmogoyfOracle = "45900b2f-f6a9-4c42-9642-008f3c1cf6dd"

// pushGraveyardCardWithTypeLine seeds a typed card in the named
// player's graveyard. Bumps the layer version directly so the next
// snapshot recomputes — Tarmogoyf's CDA reads from graveyards but
// no zone-move event fires from raw push, so the engine wouldn't
// notice without the explicit bump.
func pushGraveyardCardWithTypeLine(g *game.Game, owner uuid.UUID, name, typeLine string) {
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			if p.ID != owner {
				continue
			}
			p.Graveyard.PushTop(game.Card{
				InstanceID: uuid.New(),
				Name:       name,
				TypeLine:   typeLine,
				Owner:      owner,
				Controller: owner,
			})
		}
		g.BumpLayerVersionForTest()
	})
}

func TestTarmogoyfCDAEmptyGraveyards(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[0].ID
	goyf := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Tarmogoyf",
		TypeLine:   "Creature — Lhurgoyf",
		OracleID:   tarmogoyfOracle,
		Owner:      owner,
		Controller: owner,
	})

	if got := effectivePower(t, g, goyf); got != 0 {
		t.Errorf("Tarmogoyf power empty graveyards = %d, want 0", got)
	}
	if got := effectiveToughness(t, g, goyf); got != 1 {
		t.Errorf("Tarmogoyf toughness empty graveyards = %d, want 1 (0 + 1)", got)
	}
}

func TestTarmogoyfCDAFourTypesInGraveyards(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[0].ID

	// Seed graveyards: Creature, Instant, Sorcery, Land — 4 types.
	pushGraveyardCardWithTypeLine(g, owner, "Grizzly Bears", "Creature — Bear")
	pushGraveyardCardWithTypeLine(g, owner, "Lightning Bolt", "Instant")
	pushGraveyardCardWithTypeLine(g, g.Seats[1].ID, "Wrath of God", "Sorcery") // opponent's graveyard
	pushGraveyardCardWithTypeLine(g, g.Seats[1].ID, "Forest", "Basic Land — Forest")

	goyf := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Tarmogoyf",
		TypeLine:   "Creature — Lhurgoyf",
		OracleID:   tarmogoyfOracle,
		Owner:      owner,
		Controller: owner,
	})

	if got := effectivePower(t, g, goyf); got != 4 {
		t.Errorf("Tarmogoyf power = %d, want 4 (Creature + Instant + Sorcery + Land)", got)
	}
	if got := effectiveToughness(t, g, goyf); got != 5 {
		t.Errorf("Tarmogoyf toughness = %d, want 5", got)
	}
}

func TestTarmogoyfCDAUpdatesAfterMill(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[0].ID
	pushGraveyardCardWithTypeLine(g, owner, "Grizzly Bears", "Creature — Bear")
	pushGraveyardCardWithTypeLine(g, owner, "Lightning Bolt", "Instant")
	goyf := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Tarmogoyf",
		TypeLine:   "Creature — Lhurgoyf",
		OracleID:   tarmogoyfOracle,
		Owner:      owner,
		Controller: owner,
	})

	if got := effectivePower(t, g, goyf); got != 2 {
		t.Fatalf("Tarmogoyf power pre-mill = %d, want 2 (sanity)", got)
	}

	// Mill an enchantment into a graveyard — should bump goyf to 3/4.
	pushGraveyardCardWithTypeLine(g, owner, "Pacifism", "Enchantment — Aura")

	if got := effectivePower(t, g, goyf); got != 3 {
		t.Errorf("Tarmogoyf power post-mill = %d, want 3 (added Enchantment type)", got)
	}
	if got := effectiveToughness(t, g, goyf); got != 4 {
		t.Errorf("Tarmogoyf toughness post-mill = %d, want 4", got)
	}
}

func TestTarmogoyfCDAUnionsAcrossPlayers(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[0].ID

	// Seat 0 has a Creature; seat 1 has an Instant; seat 2 has the
	// SAME Creature type as seat 0 (should not double-count).
	pushGraveyardCardWithTypeLine(g, owner, "Grizzly Bears", "Creature — Bear")
	pushGraveyardCardWithTypeLine(g, g.Seats[1].ID, "Lightning Bolt", "Instant")
	pushGraveyardCardWithTypeLine(g, g.Seats[2].ID, "Llanowar Elves", "Creature — Elf Druid")

	goyf := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Tarmogoyf",
		TypeLine:   "Creature — Lhurgoyf",
		OracleID:   tarmogoyfOracle,
		Owner:      owner,
		Controller: owner,
	})

	if got := effectivePower(t, g, goyf); got != 2 {
		t.Errorf("Tarmogoyf power = %d, want 2 (Creature + Instant; second Creature doesn't double-count)", got)
	}
}

// TestCommanderIdentityRegressionBant verifies the layer-aware
// commanderIdentityFor still returns the commander's color
// identity correctly for a Bant ({W}{U}{G}) commander — exercises
// the S15 hand-off documented in the sprint plan.
func TestCommanderIdentityRegressionBant(t *testing.T) {
	g := newCatalogGame(t)
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
