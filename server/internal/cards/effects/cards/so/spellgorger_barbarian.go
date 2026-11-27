package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spellgorger Barbarian — Creature — Human Nightmare Barbarian {3}{R}, 3/1
// (Judgment):
//
//	"When this creature enters, discard a card at random.
//	 When this creature leaves the battlefield, draw a card."
//
// Cragganwick Cremator's random discard on the enters trigger (an empty hand
// discards nothing) and a one-card draw on the leaves trigger, which fires
// however the Barbarian left.
//
// No simplification.
func init() {
	const name = "Spellgorger Barbarian"
	effects.Register(effects.Spec{
		OracleID:     "9c0fd5ad-fbba-4598-80cf-6540888378cb",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Discards: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(name+" — discard a card at random", func(g *game.Game, item *game.StackItem) error {
				return g.DiscardRandomForEffect(item.Controller, 1)
			}),
			effects.WhenThisLeaves(name+" — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
