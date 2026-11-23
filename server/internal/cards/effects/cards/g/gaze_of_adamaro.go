package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gaze of Adamaro — Instant — Arcane {2}{R}{R} (Saviors of Kamigawa,
// #5601):
//
//	"Gaze of Adamaro deals damage to target player equal to the number of
//	 cards in that player's hand."
//
// The hand is counted as the spell resolves (CR 608.2h); a target player
// who has left the game takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91a69925-1e9b-433a-ad58-88bfd3c4bc46",
		Name:         "Gaze of Adamaro",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: player, Amount: effects.CardsInHand(ctx.Game, player)}.Apply(ctx)
		},
	})
}
