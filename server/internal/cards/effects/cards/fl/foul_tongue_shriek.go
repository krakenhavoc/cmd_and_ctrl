package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Foul-Tongue Shriek — Instant {B} (Dragons of Tarkir, #5551):
//
//	"Target opponent loses 1 life for each attacking creature you control. You gain that much life."
//
// The attackers are counted as the spell resolves, and that number is both the
// life lost and the life gained. A target that is gone fizzles the spell
// (CR 608.2b), so nothing is gained either.
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "a4dc1fea-427c-424c-99c8-af2e6d3ad2a9",
		Name:         "Foul-Tongue Shriek",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			player, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			n := len(effects.B13AttackingCreaturesYouControl(ctx.Game, ctx.Controller()))
			if n == 0 {
				return nil
			}
			if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), player, -n); err != nil {
				return err
			}
			return effects.GainLife{Amount: n}.Apply(ctx)
		},
	})
}
