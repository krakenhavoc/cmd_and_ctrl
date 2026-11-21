package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wretched Anurid — Creature — Zombie Frog Beast {1}{B}, 3/3 (Dominaria
// Remastered):
//
//	"Whenever another creature enters, you lose 1 life."
//
// Any creature entering under any player's control but the Anurid itself
// (AnotherCreatureEntered), one trigger each; the life is lost by the
// Anurid's controller, whoever's creature it was. A creature entering at
// the same time as the Anurid triggers it too (CR 603.6a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f8beeb0a-750f-4afa-a244-10011342bbeb",
		Name:         "Wretched Anurid",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AnotherCreatureEntered, "Wretched Anurid — you lose 1 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return g.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -1)
				}),
		},
	})
}
