package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// world_permanents_test.go — ADR 0109 §8 (#1862): the world enchantments
// the world rule (CR 704.5k) unblocked, and the rule itself through a
// real entry.

const (
	LivingPlaneOracle       = "30e99293-3212-4e3f-b543-b1c7c416575d"
	PillarTombsOracle       = "b8cdf79f-b247-4346-8122-9a2d7d23b3c3"
	MysticDecreeOracle      = "06eb8b90-aad8-41c1-bbc5-85ded2fe76bc"
	NetherVoidOracle        = "53bab610-1d27-4897-a4ad-cfecca82b811"
	SerraAviaryOracle       = "7a06483b-e71c-4d11-86ea-3b48e2a9cdaf"
	InTheEyeOfChaosOracle   = "6a0da9f3-cb06-42b1-ae73-b142c0eedaff"
	BazaarOfWondersOracle   = "7fca65b8-01fe-4858-97c6-f97ae82cd801"
	GravitySphereOracle     = "8ddf93fe-980b-4dc4-b56f-6a2ee50100a6"
	TeferisRealmOracle      = "9851d934-2e07-49c8-b08b-15f96d0f3f0c"
	StormWorldOracle        = "868f4ab2-a846-4ad0-8720-95fd234dd36b"
	ChaosphereOracle        = "ed01d5d7-8f34-47b3-9ca8-d82c242d38b4"
	EyeOfSingularityOracle  = "e13edf92-cdba-4c24-b81b-088b1554fde6"
	ConcordantCrossroadsOrc = "ff01b408-6d17-40a3-9efd-a1b341ec1307"
)

// wpCard is a battlefield card as the layers leave it now.
func WpCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var out game.Card
	found := false
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out, found = c, true
			}
		}
	})
	if !found {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return out
}

func WpHas(g *game.Game, c game.Card, kw string) bool {
	return effects.HasKeyword(kw)(g, uuid.Nil, c)
}

func WpOnBattlefield(g *game.Game, id uuid.UUID) bool {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}
