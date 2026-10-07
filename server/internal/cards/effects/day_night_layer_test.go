package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// day_night_layer_test.go — a static ability that reads the designation
// ("as long as it's night, ~ gets +2/+2") is a layer input that no
// permanent moving stands in for, so EventDayNightChanged has to
// invalidate the layer cache (ADR 0132). No Commander-legal card prints
// one yet, which is why the probe is a test-only spec.

const nightStaticProbeOracle = "test-night-static-probe"

func init() {
	Register(Spec{
		OracleID: nightStaticProbeOracle,
		Name:     "Night Static Probe",
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && g.IsNight()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 2
				c.Toughness += 2
			},
		}},
	})
}

func TestADayNightStaticFollowsTheDesignation(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: id, Name: "Night Static Probe", TypeLine: "Creature — Bear",
		OracleID: nightStaticProbeOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	if got := effectivePower(t, g, id); got != 1 {
		t.Fatalf("neither day nor night: power %d, want 1", got)
	}
	g.WithWriteLock(func() { g.BecomeNightForEffect() })
	if got := effectivePower(t, g, id); got != 3 {
		t.Errorf("night: power %d, want 3 — the layer cache was not invalidated by the flip", got)
	}
	g.WithWriteLock(func() { g.BecomeDayForEffect() })
	if got := effectivePower(t, g, id); got != 1 {
		t.Errorf("day: power %d, want 1", got)
	}
}
