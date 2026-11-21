package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Omen of the Sea — {1}{U} Enchantment, flash: "When this enchantment
// enters, scry 2, then draw a card. {2}{U}, Sacrifice this enchantment:
// Scry 2."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "41960d32-ddb5-42be-94b2-3a2e77ca148d",
		Name:            "Omen of the Sea",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Purpose:         game.Purpose{Scry: 2, Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Omen of the Sea — scry 2, then draw a card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					// The draw waits for the scry's answer.
					return effects.Scry{N: 2, ThenTail: ctx.TailN(effects.ThenDrawCards, 1)}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}, Sacrifice this enchantment: Scry 2.",
			Purpose: game.Purpose{Scry: 2, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.SacrificeThis()),
			Effect:  effects.Do(effects.Scry{N: 2}),
		}},
	})
}
