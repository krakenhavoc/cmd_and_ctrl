package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Get Lost — Instant {1}{W}:
//
//	"Destroy target creature, enchantment, or planeswalker. Its
//	 controller creates two Map tokens. (They're artifacts with "{1},
//	 {T}, Sacrifice this token: Target creature you control explores.
//	 Activate only as a sorcery.")"
//
// Fateful Absence's shape: read the controller before the destroy, then
// hand the victim the compensation. The Map token is CR 111.10s
// (explores.go, #2720).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "522dd417-364b-44ab-8ca9-fb55db5f26a6",
		Name:         "Get Lost",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetPermanent("target creature, enchantment, or planeswalker",
			effects.Or(effects.Creature(), effects.Enchantment(), effects.Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			controller, ok := effects.ControllerOfTarget(ctx, t.ID)
			if !ok {
				return nil
			}
			if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateToken{Controller: controller, Template: effects.MapToken(), N: 2}.Apply(ctx)
		},
	})
}
