package cardsst

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stinging Vitriol — Sorcery {B}{R} (Reality Fracture, tracker #2795):
//
//	"Stinging Vitriol deals 2 damage to target opponent. That player
//	 reveals their hand. You choose a nonland card from it. They
//	 discard that card."
//
// The damage, then the revealed-hand pick (ADR 0116) in printed order.
// A hand with no nonland card is revealed and nothing is discarded.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "624cf822-0b4e-4f0c-9893-a53cb094f616",
		Name:         "Stinging Vitriol",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		// The choice waits for a prompt the damage asks (a CR 616 order; CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				victim := effects.TargetedPlayer(ctx)
				if victim == uuid.Nil {
					return nil
				}
				return effects.DealDamage{Source: ctx.Source(), Target: victim, Amount: 2}.Apply(ctx)
			}),
			effects.Instruction(effects.ChooseNonlandFromTargetedHand),
		),
	})
}
