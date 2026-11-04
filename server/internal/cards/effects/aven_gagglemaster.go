package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "8c601628-b1ed-467b-ba92-8572343147f8",
		Name:            "Aven Gagglemaster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Aven Gagglemaster — you gain 2 life for each creature you control with flying", func(g *game.Game, item *game.StackItem) error {
				return GainLife{Amount: 2 * jmpB7CreaturesYouControlWithFlying(g, item.Controller)}.Apply(NewContext(g, item))
			}),
		},
	})
}
