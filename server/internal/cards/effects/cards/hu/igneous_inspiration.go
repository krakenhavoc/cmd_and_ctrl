package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Igneous Inspiration — Sorcery {2}{R} (EDHREC rank 25744):
//
//	"Igneous Inspiration deals 3 damage to any target.
//	 Learn. (You may reveal a Lesson card you own from outside the game
//	 and put it into your hand, or discard a card to draw a card.)"
//
// The damage can pause on a question (a replacement or prevention order,
// CR 616), so learn is a step that waits for the answer (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e31af8f-5e94-4888-857d-911253eac2cd",
		Name:         "Igneous Inspiration",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.WithRiders(effects.ForTargets(effects.DamageToTarget(0, 3)), game.Purpose{Learn: 1}),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 3}.Apply(ctx)
				}
				return nil
			}),
			effects.Learn{},
		),
	})
}
