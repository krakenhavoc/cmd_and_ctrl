package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
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
	Register(Spec{
		OracleID:     "624cf822-0b4e-4f0c-9893-a53cb094f616",
		Name:         "Stinging Vitriol",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			victim := TargetedPlayer(ctx)
			if victim == uuid.Nil {
				return nil
			}
			if err := (DealDamage{Source: ctx.Source(), Target: victim, Amount: 2}).Apply(ctx); err != nil {
				return err
			}
			return ChooseFromRevealedHand{Player: victim, Filter: Nonland(), Label: "nonland card"}.Apply(ctx)
		},
	})
}
