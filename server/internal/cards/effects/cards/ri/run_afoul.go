package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Run Afoul — Instant {G} (Core Set 2021, #5511):
//
//	"Target opponent sacrifices a creature of their choice with flying."
//
// Crude Bent Blade's targeted edict narrowed to a creature with flying;
// the opponent picks. With no flyer, nothing is sacrificed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5462cb38-7cd7-46e8-a273-35d0d1af45b4",
		Name:         "Run Afoul",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if t, ok := effects.FirstLegalTarget(ctx); ok && t.Kind == game.TargetPlayer {
				ctx.Game.PlayerSacrificesForEffect(ctx.Source(), t.ID,
					effects.SacrificeSpec("a creature with flying", effects.Creature(), effects.HasKeyword("flying")),
					"Run Afoul — sacrifice a creature with flying")
			}
			return nil
		},
	})
}
