package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goldmeadow — Plane — Lorwyn (#4870):
//
//	"Whenever a land enters, that land's controller creates three 0/1 white
//	 Goat creature tokens.
//	 Whenever chaos ensues, create a 0/1 white Goat creature token."
//
// The land ability triggers for every land of every player; the planar
// controller controls the trigger, but the tokens are the land's controller's,
// read off the land as it entered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f12400e5-8288-48e1-b3b7-0cc4f8d9ca67",
		Name:         "Goldmeadow",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, b6ALandEntered, "Goldmeadow — that land's controller creates three 0/1 white Goat creature tokens",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					who := b6TriggerObjectController(g, ctx)
					if who == uuid.Nil {
						return nil
					}
					return CreateToken{Controller: who, Template: TokenCard("0/1 white Goat"), N: 3}.Apply(ctx)
				}),
			WheneverChaosEnsues("Goldmeadow — create a 0/1 white Goat creature token",
				Do(CreateToken{Template: TokenCard("0/1 white Goat"), N: 1})),
		},
	})
}
