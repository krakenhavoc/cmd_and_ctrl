package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quill-Slinger Boggart — Creature — Goblin Warrior {3}{B}, 3/2 (Lorwyn):
//
//	"Whenever a player casts a Kithkin spell, you may have target player lose
//	 1 life."
//
// Any player's cast counts, the Boggart's controller's included
// (TmcB8APlayerCast); the Kithkin word is read through the Boggart's text
// (CR 612.2). The trigger targets a player as it goes on the stack
// (CR 603.3d), and the loss is a "you may" (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "593b229b-579b-4052-8a00-dac787b149b1",
		Name:         "Quill-Slinger Boggart",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Optional(effects.On(game.EventCast, effects.TmcB8APlayerCast(effects.OfSubtype("Kithkin")),
				"Quill-Slinger Boggart — you may have target player lose 1 life",
				func(g *game.Game, item *game.StackItem) error {
					return effects.B43TargetPlayerLoses(g, item, 1)
				}),
				"Quill-Slinger Boggart — have target player lose 1 life?"),
				effects.TargetPlayer("target player")),
		},
	})
}
