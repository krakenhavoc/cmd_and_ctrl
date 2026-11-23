package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aven Gagglemaster — Creature — Bird Warrior {3}{W}{W}, 4/3 (Jumpstart):
//
//	"Flying
//	 When this creature enters, you gain 2 life for each creature you
//	 control with flying."
//
// The fliers are counted as the trigger resolves, the Gagglemaster itself
// among them while it is still on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8c601628-b1ed-467b-ba92-8572343147f8",
		Name:            "Aven Gagglemaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Aven Gagglemaster — you gain 2 life for each creature you control with flying", func(g *game.Game, item *game.StackItem) error {
				return effects.GainLife{Amount: 2 * effects.JmpB7CreaturesYouControlWithFlying(g, item.Controller)}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
