package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zhalfirin Decoy — Creature — Human Soldier {1}{W}, 1/3 (Modern Horizons):
//
//	"{T}: Tap target creature. Activate only if you had a creature enter the
//	 battlefield under your control this turn."
//
// The window is an activation condition (CR 602.5b) on the turn's tally of
// creatures that entered under your control, judged as each entered, so one
// that has since left still counts and the Decoy itself counts the turn it
// arrives (though its own {T} still needs it free of summoning sickness, CR
// 302.6). Any creature is a legal target, yours included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b22e55ef-fb74-4991-a43b-9d3ce06c23e5",
		Name:         "Zhalfirin Decoy",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Tap target creature. Activate only if you had a creature enter the battlefield under your control this turn.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return effects.GapB193YouControlACreatureEntered(g, controller)
			},
			Effect: effects.B36TapChosenCreature,
		}},
	})
}
