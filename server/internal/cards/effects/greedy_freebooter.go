package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Greedy Freebooter — Creature — Human Pirate {B}, 1/1 (EDHREC rank
// 3221):
//
//	"When this creature dies, scry 1 and create a Treasure token."
//
// A one-mana body that pays for itself in a sacrifice deck. The dies
// trigger is the shared ThisDied condition (graveyard only — bounce
// or exile don't count); scry and the token creation have no
// ordering dependency on each other, so both run in the one Effect
// closure.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d4a9d63a-5a8a-4f51-9e49-43d6dc3234b1",
		Name:         "Greedy Freebooter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Greedy Freebooter — scry 1 and create a Treasure token",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (Scry{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
						return err
					}
					return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}.Apply(ctx)
				}),
		},
	})
}
