package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nasty End — Instant {1}{B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Draw two cards. If the sacrificed creature was legendary, draw three
//	 cards instead."
//
// "Was legendary" is read off the sacrificed creature as it last existed
// on the battlefield (CR 608.2h), from the payment record (ADR 0113 §1):
// a creature made legendary by an effect counts, and so does a
// sacrificed legendary commander.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "f73da5d8-fd15-4315-ad3c-c86c28087285",
		Name:           "Nasty End",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := 2
			if effects.SacrificedHadSupertype(ctx, "Legendary") {
				n = 3
			}
			return effects.DrawCards{N: n}.Apply(ctx)
		},
	})
}
