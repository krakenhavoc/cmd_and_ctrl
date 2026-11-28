package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sewer Rats — Creature — Rat {B}, 1/1 (Mirage, S82 OLD batch 809, #5633):
//
//	"{B}, Pay 1 life: This creature gets +1/+0 until end of turn. Activate no
//	 more than three times each turn."
//
// The limit is AtMostNActivationsPerTurn (Roterothopter's); the life is a
// cost, checked before anything is paid (CR 119.4).
//
// No simplification.
func init() {
	const label = "{B}, Pay 1 life: This creature gets +1/+0 until end of turn. Activate no more than three times each turn."
	effects.Register(effects.Spec{
		OracleID:     "bf526a0d-65cc-445f-b2b8-19a2cfdd836b",
		Name:         "Sewer Rats",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.Plus(effects.ManaCost("{B}"), effects.PayLife(1)),
			Condition: effects.AtMostNActivationsPerTurn(label, 3),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Sewer Rats — +1/+0"}.Apply(ctx)
			},
		}},
	})
}
