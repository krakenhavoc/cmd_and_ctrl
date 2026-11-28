package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Liege of the Hollows — Creature — Spirit {2}{G}{G}, 3/4 (Weatherlight, S82 older sets batch 780, #5629):
//
//	"When this creature dies, each player may pay any amount of mana. Then each player creates a number of 1/1 green Squirrel
//	 creature tokens equal to the amount of mana they paid this way."
//
// Starting with the controller, each player in turn order is asked how much of their own mana to pay (any amount, none
// included, CR 608.2d) and then creates that many Squirrels under their own control (each player's own count, not a shared total, so
// not join forces). The tokens are made as each answer lands, which no one can tell from making them after the last: nothing in
// between depends on them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bef7441e-5a44-46bc-8141-8e02ecd06fce",
		Name:         "Liege of the Hollows",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Liege of the Hollows — each player may pay any amount of mana, then creates that many 1/1 green Squirrels",
				func(g *game.Game, item *game.StackItem) error {
					return liegeOfTheHollows.Start(effects.NewContext(g, item))
				}),
		},
	})
}

var liegeOfTheHollows = effects.NewS82OldB780LiegeOfTheHollows("pay-unless/s82-old-b780-liege-of-the-hollows")
