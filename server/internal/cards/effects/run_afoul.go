package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Run Afoul — Instant {G} (Core Set 2021, #5511):
//
//	"Target opponent sacrifices a creature of their choice with flying."
//
// Crude Bent Blade's targeted edict narrowed to a creature with flying;
// the opponent picks. With no flyer, nothing is sacrificed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5462cb38-7cd7-46e8-a273-35d0d1af45b4",
		Name:         "Run Afoul",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if t, ok := firstLegalTarget(ctx); ok && t.Kind == game.TargetPlayer {
				ctx.Game.PlayerSacrificesForEffect(ctx.Source(), t.ID,
					sacrificeSpec("a creature with flying", Creature(), HasKeyword("flying")),
					"Run Afoul — sacrifice a creature with flying")
			}
			return nil
		},
	})
}
