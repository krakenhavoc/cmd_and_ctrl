package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Argent Sphinx — Creature — Sphinx {2}{U}{U}, 4/3 (Scars of Mirrodin, S82
// OLD batch 469, #5577):
//
//	"Flying
//	 Metalcraft — {U}: Exile this creature. Return it to the battlefield under
//	 your control at the beginning of the next end step. Activate only if you
//	 control three or more artifacts."
//
// Aetherling's flicker (ExileThisThenScheduleReturn) gated as it is activated
// (CR 602.5b). It comes back as a new object at the next end step (CR 400.7),
// under its owner's control, which is the activator's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7220c391-8d0e-4f43-94fc-ff77ffa2ee8d",
		Name:            "Argent Sphinx",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:     "Metalcraft — {U}: Exile this creature. Return it to the battlefield under your control at the beginning of the next end step. Activate only if you control three or more artifacts.",
			Purpose:   game.Purpose{Answers: game.AnswerSurvive},
			Cost:      effects.ManaCost("{U}"),
			Condition: effects.OldB469MetalcraftActivation(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ExileThisThenScheduleReturn(effects.NewContext(g, item), item, "Argent Sphinx — return the exiled creature")
			},
		}},
	})
}
