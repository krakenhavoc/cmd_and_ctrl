package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Impact Resonance — Instant {1}{R} (Commander 2014):
//
//	"Impact Resonance deals X damage divided as you choose among any number
//	 of target creatures, where X is the greatest amount of damage dealt by
//	 a source to a permanent or player this turn."
//
// X is fixed as the spell is cast, with the division (CR 601.2d); with no
// damage dealt yet there is nothing to divide and no target can be chosen.
// Each target is judged again on resolution and a share is not
// redistributed (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "94846856-fe68-4ada-8c34-a54de568e803",
		Name:         "Impact Resonance",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("any number of target creatures").WithCount(0, 0).Dividing(effects.DivideBy(effects.DivideGreatestDamageDealtThisTurn)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DealDividedDamage(ctx)
		},
	})
}
