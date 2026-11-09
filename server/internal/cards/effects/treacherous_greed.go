package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Treacherous Greed — Instant {1}{W}{B} (Murders at Karlov Manor, #4469):
//
//	"As an additional cost to cast this spell, sacrifice a creature that
//	 dealt damage this turn.
//	 Draw three cards. Each opponent loses 3 life and you gain 3 life."
//
// The cost's clause is a predicate over the creature (the source of a
// damage event this turn, DealtDamageAsASourceThisTurn): any creature you
// control that dealt combat or noncombat damage, whether or not what it hit
// is still around. A cast with no such creature is refused at announce.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "5483f787-7aca-4efe-a39a-ea29f8c07f0d",
		Name:           "Treacherous Greed",
		Completeness:   CompletenessFull,
		Purpose:        game.Purpose{Draws: 3, LifeGain: 3},
		AdditionalCost: SacrificeCost("a creature that dealt damage this turn", Creature(), DealtDamageAsASourceThisTurn()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return Do(DrawCards{N: 3}, EachOpponentLosesLife{N: 3}, GainLife{Amount: 3})(ctx.Game, ctx.Item)
		},
	})
}
