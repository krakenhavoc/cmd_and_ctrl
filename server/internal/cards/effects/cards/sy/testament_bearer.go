package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Testament Bearer — Creature — Phyrexian Warrior {3}{B}, 4/1:
//
//	"When this creature dies, look at the top three cards of your
//	 library. Put one of them into your hand and the rest into your
//	 graveyard."
//
// Wail of the Forgotten's look: the pick is mandatory (one card, floor one),
// and the rest go to the graveyard. A library of fewer than three looks at
// what it has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce0f6cd6-35d5-4f52-b81e-6d15e3bf916e",
		Name:         "Testament Bearer",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Testament Bearer — look at the top three, one to hand, rest to graveyard", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.TakeFromLibraryToHand{
					Player: item.Controller,
					Cards:  g.LookAtTopOfLibraryForEffect(item.Controller, 3),
					Max:    1,
					Label:  "Testament Bearer — put one into your hand",
					Key:    effects.TakeRestGraveyard,
				}.Apply(ctx)
			}),
		},
	})
}
