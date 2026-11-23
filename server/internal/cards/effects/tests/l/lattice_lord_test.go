package testsl

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestMycosynthLatticeAddsArtifact: a Forest in play picks up the
// Artifact type when Mycosynth Lattice resolves.
func TestMycosynthLatticeAddsArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[0].ID
	forest := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Forest",
		TypeLine:   "Basic Land — Forest",
		Owner:      owner,
		Controller: owner,
	})

	// Sanity: before lattice, Forest is just Land.
	if got := effectstest.EffectiveTypes(t, g, forest); len(got) != 1 || got[0] != "Land" {
		t.Fatalf("Forest pre-lattice types = %v, want [Land]", got)
	}

	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Mycosynth Lattice",
		TypeLine:   "Artifact",
		OracleID:   effectstest.MycosynthLatticeOracle,
		Owner:      owner,
		Controller: owner,
	})

	got := effectstest.EffectiveTypes(t, g, forest)
	if !effectstest.ContainsString(got, "Land") {
		t.Errorf("Forest post-lattice should still be Land; got %v", got)
	}
	if !effectstest.ContainsString(got, "Artifact") {
		t.Errorf("Forest post-lattice should be Artifact; got %v", got)
	}
}

// TestMycosynthLatticeIdempotent: pre-existing artifacts don't
// double-stamp.
func TestMycosynthLatticeIdempotent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[0].ID
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Some Artifact",
		TypeLine:   "Artifact",
		Owner:      owner,
		Controller: owner,
	})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Mycosynth Lattice",
		TypeLine:   "Artifact",
		OracleID:   effectstest.MycosynthLatticeOracle,
		Owner:      owner,
		Controller: owner,
	})

	got := effectstest.EffectiveTypes(t, g, rock)
	count := 0
	for _, t := range got {
		if t == "Artifact" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("Artifact present %d times in %v, want 1 (idempotent stamp)", count, got)
	}
}

// TestLordOfAtlantisGrantsToOtherMerfolkOnly: a Merfolk gets the
// +1/+1 + islandwalk; a non-Merfolk creature stays printed; the
// Lord itself is excluded by the "other" predicate.
func TestLordOfAtlantisGrantsToOtherMerfolkOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[0].ID

	merfolk := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Llanowar Merfolk",
		TypeLine:   "Creature — Merfolk",
		Power:      1,
		Toughness:  1,
		Owner:      owner,
		Controller: owner,
	})
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Grizzly Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      owner,
		Controller: owner,
	})
	lord := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Lord of Atlantis",
		TypeLine:   "Creature — Merfolk",
		Power:      2,
		Toughness:  2,
		OracleID:   effectstest.LordOfAtlantisOracle,
		Owner:      owner,
		Controller: owner,
	})

	// Merfolk gets +1/+1 + islandwalk.
	if got := effectstest.EffectivePower(t, g, merfolk); got != 2 {
		t.Errorf("Merfolk power = %d, want 2 (printed 1 + lord 1)", got)
	}
	if got := effectstest.EffectiveAbilities(t, g, merfolk); !effectstest.ContainsString(got, "islandwalk") {
		t.Errorf("Merfolk abilities = %v, want islandwalk", got)
	}

	// Bear (non-Merfolk) unchanged.
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("Bear power = %d, want 2 (printed; no lord buff)", got)
	}
	if got := effectstest.EffectiveAbilities(t, g, bear); effectstest.ContainsString(got, "islandwalk") {
		t.Errorf("Bear should not have islandwalk; got %v", got)
	}

	// Lord excludes itself ("other" predicate).
	if got := effectstest.EffectivePower(t, g, lord); got != 2 {
		t.Errorf("Lord self-power = %d, want 2 (printed; 'other' predicate excludes self)", got)
	}
	if got := effectstest.EffectiveAbilities(t, g, lord); effectstest.ContainsString(got, "islandwalk") {
		t.Errorf("Lord should not grant islandwalk to itself; got %v", got)
	}
}

// TestTwoLordsBuffEachOther: two Lords on the battlefield each
// pump the OTHER (not themselves), so each one is 3/3 with
// islandwalk in the wire.
func TestTwoLordsBuffEachOther(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[0].ID

	lordA := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Lord A",
		TypeLine:   "Creature — Merfolk",
		Power:      2,
		Toughness:  2,
		OracleID:   effectstest.LordOfAtlantisOracle,
		Owner:      owner,
		Controller: owner,
	})
	lordB := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Lord B",
		TypeLine:   "Creature — Merfolk",
		Power:      2,
		Toughness:  2,
		OracleID:   effectstest.LordOfAtlantisOracle,
		Owner:      owner,
		Controller: owner,
	})

	for _, id := range []uuid.UUID{lordA, lordB} {
		if got := effectstest.EffectivePower(t, g, id); got != 3 {
			t.Errorf("Lord power = %d, want 3 (each pumps the other)", got)
		}
		if got := effectstest.EffectiveAbilities(t, g, id); !effectstest.ContainsString(got, "islandwalk") {
			t.Errorf("Lord abilities = %v, want islandwalk (each grants to the other)", got)
		}
	}
}
