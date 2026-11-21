package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loyal Gryff — Creature — Hippogriff {2}{W}, 2/2 (Innistrad: Midnight Hunt, S82):
//
//	"Flash
//	 Flying
//	 When this creature enters, you may return another creature you control to its owner's hand."
//
// First Responder's pick: nothing targets, the creature is chosen as it resolves, zero being the "you may" decline, and the Gryff itself is never offered (s82MidB01Return).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f5c07218-c8bf-4990-a366-f3c065d0b30c",
		Name:            "Loyal Gryff",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Loyal Gryff — you may return another creature you control to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ChooseAnotherCreatureYouControl(g, item,
						"Loyal Gryff — return another creature you control to its owner's hand?", effects.S82MidB01Return)
				}),
		},
	})
}
