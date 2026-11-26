package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vampire Lacerator — Creature — Vampire Warrior {B}, 2/2 (Zendikar, S82 OLD b495):
//
//	"At the beginning of your upkeep, you lose 1 life unless an opponent has 10 or
//	 less life."
//
// An upkeep trigger whose "unless" is read as it resolves: if some opponent still
// in the game has 10 or less life, nothing is lost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dc261758-0982-4b7d-9986-042c03cc8366",
		Name:         "Vampire Lacerator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Vampire Lacerator — you lose 1 life unless an opponent has 10 or less life",
				func(g *game.Game, item *game.StackItem) error {
					if effects.Old495AnOpponentHasLifeAtMost(g, item.Controller, 10) {
						return nil
					}
					return effects.GainLife{Amount: -1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
