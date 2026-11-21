package effectstest

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

const MycosynthLatticeOracle = "ae1f2ab5-c6a5-4d49-a746-3cb4668bf805"
const LordOfAtlantisOracle = "cc7f290f-ca00-4285-9bdb-4b4402444f30"

// effectiveTypes returns Effective().Types for the named card on
// the battlefield, forcing a snapshot to ensure the layer engine
// has resolved.
func EffectiveTypes(t *testing.T, g *game.Game, cardID uuid.UUID) []string {
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
func EffectiveAbilities(t *testing.T, g *game.Game, cardID uuid.UUID) []string {
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

func ContainsString(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}
