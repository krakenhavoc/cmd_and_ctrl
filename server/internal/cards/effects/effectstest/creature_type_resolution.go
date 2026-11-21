package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// creature_type_resolution_test.go — #2382. A creature-type choice made
// as a spell resolves: the prompt is the as-enters one, the rest of the
// card hangs off the answer.

const (
	DistantMelodyOracle     = "ac6c5852-71c7-4f19-8c87-ccb345052862"
	KindredDominanceOracle  = "ccaa44f2-96be-44e2-884f-c31baa3908d5"
	RaiseThePalisadeOracle  = "f55a3781-fe33-4301-9bb5-6a54b9c13c4f"
	BannerOfKinshipOracle   = "8c220dbd-6572-4715-aae6-dd09a4252d68"
	PatriarchsBiddingOracle = "25fc3bc2-d852-4f52-9adb-c7e35c06f3af"
	TribalUnityOracle       = "6b4ee935-4474-4757-9fab-978a2e1e72a0"
	KindredChargeOracle     = "eb0b5b5f-54d3-49c6-9ad8-d44a5551acdb"
)

// answerCreatureType answers the open creature-type prompt owed by
// `chooser`, failing if none is open.
func AnswerCreatureType(t *testing.T, g *game.Game, chooser uuid.UUID, tribe string) {
	t.Helper()
	c := PendingOfKind(g, game.PendingChoiceCreatureType)
	if c == nil {
		t.Fatalf("no creature-type prompt is open (answering %s)", tribe)
	}
	if c.Chooser != chooser {
		t.Fatalf("the open creature-type prompt belongs to %s, not %s", c.Chooser, chooser)
	}
	if err := g.ResolveCreatureTypeChoice(c.ID, chooser, tribe); err != nil {
		t.Fatalf("ResolveCreatureTypeChoice(%s): %v", tribe, err)
	}
}

// castAndPause casts the spell, passes priority until it resolves, and
// returns with the creature-type prompt open.
func CastAndPause(t *testing.T, g *game.Game, name, typeLine, oracle string) {
	t.Helper()
	CastCatalogSpell(t, g, name, typeLine, oracle, nil)
	PassPriorityAroundTable(t, g)
	if PendingOfKind(g, game.PendingChoiceCreatureType) == nil {
		t.Fatalf("%s: resolving did not pause on a creature-type prompt", name)
	}
}

func HasKw(kws []string, kw string) bool {
	for _, k := range kws {
		if k == kw {
			return true
		}
	}
	return false
}

func BattlefieldHas(g *game.Game, id uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() { found = g.Battlefield.Contains(id) })
	return found
}

func InHandOf(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Hand.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func BattlefieldHasName(g *game.Game, name string) bool {
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				found = true
			}
		}
	})
	return found
}
