package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Melt Terrain — Sorcery {2}{R}{R} (Scars of Mirrodin, S82 OLD batch 471,
// #5577):
//
//	"Destroy target land. Melt Terrain deals 2 damage to that land's
//	 controller."
//
// The controller is read before the land is destroyed (CR 608.2h), and the
// damage is dealt whether or not the land was destroyed (an indestructible one
// still costs its controller 2). A target that is gone makes the spell do
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c0940d6-9f0d-4d84-ab28-ff6595254675",
		Name:         "Melt Terrain",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			controller, ok := effects.ControllerOfTarget(ctx, id)
			if !ok {
				return nil
			}
			if err := (effects.DestroyTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return effects.DealDamage{Source: ctx.Source(), Target: controller, Amount: 2}.Apply(ctx)
		},
	})
}
