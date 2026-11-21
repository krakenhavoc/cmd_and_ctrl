package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// divided_1657_test.go — #1657, ADR 0065's 2026-09-28 (b) amendment: a
// divided amount read off the board or off a paid cost, fixed at
// announce, and "up to" division. Ureni, the Song Unending (lands you
// control), Orca, Siege Demon (its last-known power), Lathiel, the
// Bounteous Dawn (life gained this turn, "up to"), and Avacyn's
// Judgment (2, or X if its madness cost was paid).

const (
	d1657UreniOracle   = "1c995c80-3301-409b-822b-9297aa260823"
	D1657OrcaOracle    = "b8600894-3229-420e-9464-d92cf0a7d0fc"
	D1657AvacynsOracle = "f3ae58ed-8ef7-4e0a-945f-1f622157236b"
)

// d1657UreniEnters casts Ureni with `lands` lands under the caster and
// returns the open pick_target prompt.
func D1657UreniEnters(t *testing.T, g *game.Game, lands int) *game.PendingChoice {
	t.Helper()
	me := g.Seats[0]
	AcLands(g, me.ID, lands, "Forest")
	CastCatalogSpell(t, g, "Ureni, the Song Unending", "Legendary Creature — Spirit Dragon", d1657UreniOracle, nil)
	PassPriorityAroundTable(t, g)
	p := LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("Ureni's enters trigger asks for targets")
	}
	return p
}

// d1657LandCannonOracle is a test-only card: no catalogued ACTIVATED
// ability divides a ruled amount yet (Polukranos waits on
// monstrosity), so the activation path's binding is proved on this.
const D1657LandCannonOracle = "test-1657-land-cannon"

func init() {
	effects.Register(effects.Spec{
		OracleID: D1657LandCannonOracle,
		Name:     "Test Land Cannon",
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This deals damage equal to the number of lands you control divided as you choose among any number of targets.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny().WithCount(0, 0).
				Dividing(effects.DivideBy(effects.DivideLandsYouControl)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DealDividedDamage(effects.NewContext(g, item))
			},
		}},
	})
}
