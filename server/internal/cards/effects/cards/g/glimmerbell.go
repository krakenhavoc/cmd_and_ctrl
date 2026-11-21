package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glimmerbell — Creature — Elemental Jellyfish {1}{U}, 1/3
// (Foundations Jumpstart):
//
//	"Flying
//	 {1}{U}: Untap this creature."
//
// Flying is a printed keyword; the untap is UntapTarget on the source,
// which does nothing once the source has left (CR 400.7). No
// simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1a8b7792-e105-40c8-b1b4-ab1a8f3731d8",
		Name:            "Glimmerbell",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			UntapSelf: true,
			Label:     "{1}{U}: Untap this creature.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.ManaCost("{1}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.UntapTarget{Target: ctx.Source()}.Apply(ctx)
			},
		}},
	})
}
