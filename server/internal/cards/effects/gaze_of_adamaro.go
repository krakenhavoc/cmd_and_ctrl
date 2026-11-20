package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "91a69925-1e9b-433a-ad58-88bfd3c4bc46",
		Name:         "Gaze of Adamaro",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player, ok := firstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			return DealDamage{Source: ctx.Source(), Target: player, Amount: cardsInHand(ctx.Game, player)}.Apply(ctx)
		},
	})
}
