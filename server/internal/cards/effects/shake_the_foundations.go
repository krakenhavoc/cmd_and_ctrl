package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shake the Foundations — Instant {2}{R} (Rivals of Ixalan, #5531):
//
//	"Shake the Foundations deals 1 damage to each creature without flying.
//	 Draw a card."
//
// Tremor's sweep, then a card: the creatures without flying are read
// through the layers as it resolves, so one granted flying is spared, and
// no player is dealt damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "99eb5cda-684d-4f6c-bc34-1f03c8797150",
		Name:     "Shake the Foundations",
		// ADR 0126 §6: a creature with flying is spared.
		Purpose:      game.Purpose{Draws: 1, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}},
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := damageEachMatching(ctx, And(Creature(), WithoutKeyword("flying")), 1); err != nil {
				return err
			}
			return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
