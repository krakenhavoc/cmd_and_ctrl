package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Omen of the Sea — {1}{U} Enchantment, flash: "When this enchantment
// enters, scry 2, then draw a card. {2}{U}, Sacrifice this enchantment:
// Scry 2."
func init() {
	Register(Spec{
		OracleID:        "41960d32-ddb5-42be-94b2-3a2e77ca148d",
		Name:            "Omen of the Sea",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Purpose:         game.Purpose{Scry: 2, Draws: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Omen of the Sea — scry 2, then draw a card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					// The draw waits for the scry's answer.
					return Scry{N: 2, ThenTail: ctx.TailN(thenDrawCards, 1)}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}, Sacrifice this enchantment: Scry 2.",
			Purpose: game.Purpose{Scry: 2, Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}{U}"), SacrificeThis()),
			Effect:  Do(Scry{N: 2}),
		}},
	})
}
