package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dramatic Rescue — Instant {W}{U} (Return to Ravnica, #5566):
//
//	"Return target creature to its owner's hand. You gain 2 life."
//
// One target: when it is illegal on resolution the whole spell does nothing
// (CR 608.2b), so no life is gained.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5386844-ea24-48ee-9337-ae10bcfddfb9",
		Name:         "Dramatic Rescue",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 2},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.BounceToHand{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Amount: 2}.Apply(ctx)
		},
	})
}
