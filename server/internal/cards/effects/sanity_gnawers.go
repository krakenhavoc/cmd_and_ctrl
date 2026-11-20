package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sanity Gnawers — Creature — Rat {1}{B}{R}, 1/1 (Alara Reborn, #5586):
//
//	"When this creature enters, target player discards a card at random."
//
// The player is targeted as the trigger goes on the stack (CR 603.3d);
// the discard is the at-random primitive (CR 701.9b), so nobody chooses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e82308d5-29ee-48eb-9d1c-f24897ef191c",
		Name:         "Sanity Gnawers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Sanity Gnawers — target player discards a card at random",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					player, ok := firstLegalPlayerTarget(ctx)
					if !ok {
						return nil
					}
					return DiscardCards{Player: player, N: 1}.Apply(ctx)
				}),
				TargetPlayer("target player")),
		},
	})
}
