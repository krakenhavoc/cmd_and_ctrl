package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Infernal Harvest — Sorcery {1}{B} (Visions, S82 OLD batch 786, #5631):
//
//	"As an additional cost to cast this spell, return X Swamps you control
//	 to their owner's hand.
//	 Infernal Harvest deals X damage divided as you choose among any number
//	 of target creatures."
//
// X is the one number the cast announces through the additional cost
// (#6122, ReturnXToHandCost); the division is the existing DivideX clause,
// so X also caps how many creatures can be chosen. A creature that has gone
// by resolution takes nothing and its share is lost (CR 608.2b); the Swamps
// stay returned.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "b8c8cde9-34db-4185-86f5-4289969880c5",
		Name:           "Infernal Harvest",
		Completeness:   effects.CompletenessFull,
		XMatters:       true,
		AdditionalCost: effects.ReturnXToHandCost("X Swamps you control", effects.HasSubtype("Swamp")),
		Targets:        effects.TargetCreature("any number of target creatures").WithCount(0, 0).Dividing(effects.DivideX()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DealDividedDamage(ctx)
		},
	})
}
