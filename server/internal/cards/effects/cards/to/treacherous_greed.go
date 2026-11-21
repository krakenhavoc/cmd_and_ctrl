package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treacherous Greed — Instant {1}{W}{B} (Murders at Karlov Manor, #4469):
//
//	"As an additional cost to cast this spell, sacrifice a creature that
//	 dealt damage this turn.
//	 Draw three cards. Each opponent loses 3 life and you gain 3 life."
//
// The cost's clause is a predicate over the creature (the source of a
// damage event this turn, mshA6DealtDamageThisTurn): any creature you
// control that dealt combat or noncombat damage, whether or not what it hit
// is still around. A cast with no such creature is refused at announce.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "5483f787-7aca-4efe-a39a-ea29f8c07f0d",
		Name:           "Treacherous Greed",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Draws: 3, LifeGain: 3},
		AdditionalCost: effects.SacrificeCost("a creature that dealt damage this turn", effects.Creature(), effects.MshA6DealtDamageThisTurn()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.Do(effects.DrawCards{N: 3}, effects.EachOpponentLosesLife{N: 3}, effects.GainLife{Amount: 3})(ctx.Game, ctx.Item)
		},
	})
}
