package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seismic Rupture — Sorcery {2}{R} (Foundations, #4465):
//
//	"Seismic Rupture deals 2 damage to each creature without flying."
//
// Subterranean Shambler's sweep at 2, on a sorcery: the creatures
// without flying are read post-layer as it resolves, so one granted
// flying is spared, and no player is dealt damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "3b1f1a57-d83a-443f-b190-03df63d2bfe6",
		Name:     "Seismic Rupture",
		// ADR 0126 §6: a creature with flying is spared.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, Partial: true}},
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx, And(Creature(), WithoutKeyword("flying")), 2)
		},
	})
}
