package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vedalken Dismisser — Creature — Vedalken Wizard {5}{U}, 2/2 (Ravnica: City of
// Guilds, S82 OLD batch 576, #5600):
//
//	"When this creature enters, put target creature on top of its owner's
//	 library."
//
// Banishment Decree's tuck: the creature is chosen as the trigger goes on the
// stack and put on top of its owner's library as it resolves if it is still
// legal. The Dismisser may target itself or any creature, and a token put into
// a library ceases to exist (CR 111.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6ed8820-4276-4ead-ba82-4654a92c415d",
		Name:         "Vedalken Dismisser",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Vedalken Dismisser — put target creature on top of its owner's library",
				func(g *game.Game, item *game.StackItem) error {
					return effects.PutTargetOnTopOfOwnersLibrary(item, effects.NewContext(g, item))
				}),
				effects.TargetCreature("target creature")),
		},
	})
}
