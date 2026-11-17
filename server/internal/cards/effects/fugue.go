package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fugue — Sorcery {3}{B}{B} (Exodus, #5626):
//
//	"Target player discards three cards."
//
// Mind Rot with three: the shared targetPlayerDiscards body (Ana
// Battlemage's three) opens a discard prompt for the target, who picks
// the cards (CR 701.9a). A player with fewer than three cards discards
// their hand. A target that left the game is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "724c8a20-bfb3-4c63-8cfa-5ad4c72ebe2f",
		Name:         "Fugue",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return targetPlayerDiscards(3)(ctx.Game, item)
		},
	})
}
