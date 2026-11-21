package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thought Scour — Instant {U} (EDHREC rank 1785):
//
//	"Target player mills two cards.
//	 Draw a card."
//
// The one-mana self-mill cantrip. Glimpse the Unthinkable's target
// clause, a mill of two, then the caster draws — in the printed
// order, so a Thought Scour aimed at yourself mills before it draws
// and the card drawn is the third from the top.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "83101ba8-a569-4827-8c53-9ca0dfcd59a7",
		Name:         "Thought Scour",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: effects.ResolveSteps(
			// The draw waits for a prompt the mill asks (CR 608.2c, #4155).
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					if err := (effects.MillCards{Player: t.ID, N: 2}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
			effects.DrawCards{N: 1},
		),
	})
}
