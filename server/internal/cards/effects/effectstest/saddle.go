package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// saddle_test.go — #2695, CR 702.171. Saddle is crew's cost over OTHER
// creatures at sorcery speed, and the designation it sets is the one
// ADR 0071 gate whose designation lasts a turn. The tests below are
// about the differences from crew: the Mount never pays for itself,
// the timing, the lifetime, and the "attacks while saddled" reading.

const (
	GildedGhodaOracle         = "5f4939dc-6e87-47fa-8287-9da60d4b5db2"
	DroverGrizzlyOracle       = "290fe09e-04ab-48d1-a60e-db47ea8c4d00"
	StubbornBurrowfiendOracle = "24bc7bd5-f262-45a2-8e94-377fef2ba5d5"
	SaddleStaticProbeOracle   = "test-saddle-static-probe"
)

func init() {
	// A Mount whose only text is "As long as this is saddled, it has
	// flying and trample", to prove the gate and its expiry without a
	// real card's other abilities in the way.
	effects.Register(effects.Spec{
		OracleID: SaddleStaticProbeOracle,
		Name:     "Saddle Static Probe",
		Static:   []game.StaticAbility{effects.SaddledKeywords("flying", "trample")},
	})
}

// pushMountForTest puts a catalog Mount on the battlefield, already
// past summoning sickness.
func PushMountForTest(g *game.Game, owner uuid.UUID, name, oracle string, power int) uuid.UUID {
	return PushDiesCreatureForTest(g, owner, name, oracle, "Creature — Horse Mount", power, power)
}

func SaddleAbility(t *testing.T, g *game.Game, owner, mount uuid.UUID, saddlers ...uuid.UUID) error {
	t.Helper()
	// The saddle ability is wherever the card prints it: after a Mount's
	// other abilities, ahead of them on none of the proof cards.
	index := 0
	if c, ok := BattlefieldCardByID(g, mount); ok {
		if spec, ok := effects.Lookup(c.OracleID); ok {
			for i, a := range spec.Activated {
				if a.Cost.Saddle > 0 {
					index = i
				}
			}
		}
	}
	return g.ActivateCatalogAbility(owner, mount, index, game.ActivateAbilityParams{CrewIDs: saddlers})
}

func SaddleCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	c, ok := BattlefieldCardByID(g, id)
	if !ok {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return c
}
