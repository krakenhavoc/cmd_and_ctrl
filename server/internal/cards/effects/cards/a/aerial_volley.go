package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aerial Volley — Instant {G} (Magic Origins):
//
//	"Aerial Volley deals 3 damage divided as you choose among one, two, or
//	 three target creatures with flying."
//
// Arc Lightning's divided shape with the target clause narrowed to
// creatures with flying: one to three targets, each getting at least 1
// of the 3 (CR 601.2d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ab855992-34da-42a2-b7c7-dec4db849159",
		Name:         "Aerial Volley",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCreature("one, two, or three target creatures with flying",
			effects.HasKeyword("flying")).WithCount(1, 3).Dividing(effects.Divide(3)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DealDividedDamage(ctx)
		},
	})
}
