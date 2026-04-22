package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lattice_lord_test.go covers the S16 sub-PR 4 catalog additions:
// Mycosynth Lattice (Layer 4 type-add) and Lord of Atlantis
// (Layer 6 ability grant + Layer 7c +1/+1 with self-exclusion).
//
// Uses the layer-aware battlefield-push helper from anthem_test.go
// so EnteredBattlefieldAt + layerVersion are stamped properly via
// the listener.

const mycosynthLatticeOracle = "16ddc91a-4b7c-4d12-bd0a-1b2f4ce5d12e"
const lordOfAtlantisOracle = "9e64b11a-eaca-4ce4-a76a-29f4f12c8c6f"

// effectiveTypes returns Effective().Types for the named card on
// the battlefield, forcing a snapshot to ensure the layer engine
// has resolved.
func effectiveTypes(t *testing.T, g *game.Game, cardID uuid.UUID) []string {
	t.Helper()
	var types []string
	var found bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != cardID {
				continue
			}
			types = c.Effective().Types
			found = true
			return
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", cardID)
	}
	return types
}

// effectiveAbilities returns Effective().Abilities for the named
// card. Forces a snapshot.
func effectiveAbilities(t *testing.T, g *game.Game, cardID uuid.UUID) []string {
	t.Helper()
	var abilities []string
	var found bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != cardID {
				continue
			}
			abilities = c.Effective().Abilities
			found = true
			return
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", cardID)
	}
	return abilities
}

func containsString(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}

// TestMycosynthLatticeAddsArtifact: a Forest in play picks up the
// Artifact type when Mycosynth Lattice resolves.
func TestMycosynthLatticeAddsArtifact(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[0].ID
	forest := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Forest",
		TypeLine:   "Basic Land — Forest",
		Owner:      owner,
		Controller: owner,
	})

	// Sanity: before lattice, Forest is just Land.
	if got := effectiveTypes(t, g, forest); len(got) != 1 || got[0] != "Land" {
		t.Fatalf("Forest pre-lattice types = %v, want [Land]", got)
	}

	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Mycosynth Lattice",
		TypeLine:   "Artifact",
		OracleID:   mycosynthLatticeOracle,
		Owner:      owner,
		Controller: owner,
	})

	got := effectiveTypes(t, g, forest)
	if !containsString(got, "Land") {
		t.Errorf("Forest post-lattice should still be Land; got %v", got)
	}
	if !containsString(got, "Artifact") {
		t.Errorf("Forest post-lattice should be Artifact; got %v", got)
	}
}

// TestMycosynthLatticeIdempotent: pre-existing artifacts don't
// double-stamp.
func TestMycosynthLatticeIdempotent(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[0].ID
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Some Artifact",
		TypeLine:   "Artifact",
		Owner:      owner,
		Controller: owner,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Mycosynth Lattice",
		TypeLine:   "Artifact",
		OracleID:   mycosynthLatticeOracle,
		Owner:      owner,
		Controller: owner,
	})

	got := effectiveTypes(t, g, rock)
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
	g := newCatalogGame(t)
	owner := g.Seats[0].ID

	merfolk := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Llanowar Merfolk",
		TypeLine:   "Creature — Merfolk",
		Power:      1,
		Toughness:  1,
		Owner:      owner,
		Controller: owner,
	})
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Grizzly Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      owner,
		Controller: owner,
	})
	lord := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Lord of Atlantis",
		TypeLine:   "Creature — Merfolk",
		Power:      2,
		Toughness:  2,
		OracleID:   lordOfAtlantisOracle,
		Owner:      owner,
		Controller: owner,
	})

	// Merfolk gets +1/+1 + islandwalk.
	if got := effectivePower(t, g, merfolk); got != 2 {
		t.Errorf("Merfolk power = %d, want 2 (printed 1 + lord 1)", got)
	}
	if got := effectiveAbilities(t, g, merfolk); !containsString(got, "islandwalk") {
		t.Errorf("Merfolk abilities = %v, want islandwalk", got)
	}

	// Bear (non-Merfolk) unchanged.
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("Bear power = %d, want 2 (printed; no lord buff)", got)
	}
	if got := effectiveAbilities(t, g, bear); containsString(got, "islandwalk") {
		t.Errorf("Bear should not have islandwalk; got %v", got)
	}

	// Lord excludes itself ("other" predicate).
	if got := effectivePower(t, g, lord); got != 2 {
		t.Errorf("Lord self-power = %d, want 2 (printed; 'other' predicate excludes self)", got)
	}
	if got := effectiveAbilities(t, g, lord); containsString(got, "islandwalk") {
		t.Errorf("Lord should not grant islandwalk to itself; got %v", got)
	}
}

// TestTwoLordsBuffEachOther: two Lords on the battlefield each
// pump the OTHER (not themselves), so each one is 3/3 with
// islandwalk in the wire.
func TestTwoLordsBuffEachOther(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[0].ID

	lordA := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Lord A",
		TypeLine:   "Creature — Merfolk",
		Power:      2,
		Toughness:  2,
		OracleID:   lordOfAtlantisOracle,
		Owner:      owner,
		Controller: owner,
	})
	lordB := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Lord B",
		TypeLine:   "Creature — Merfolk",
		Power:      2,
		Toughness:  2,
		OracleID:   lordOfAtlantisOracle,
		Owner:      owner,
		Controller: owner,
	})

	for _, id := range []uuid.UUID{lordA, lordB} {
		if got := effectivePower(t, g, id); got != 3 {
			t.Errorf("Lord power = %d, want 3 (each pumps the other)", got)
		}
		if got := effectiveAbilities(t, g, id); !containsString(got, "islandwalk") {
			t.Errorf("Lord abilities = %v, want islandwalk (each grants to the other)", got)
		}
	}
}
