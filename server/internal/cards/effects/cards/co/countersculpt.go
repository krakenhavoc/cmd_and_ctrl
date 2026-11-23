package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Countersculpt — Instant {U}{U} (Reality Fracture):
//
//	"As an additional cost to cast this spell, behold a Jace or pay
//	 {1}. (To behold a Jace, choose a Jace you control or reveal a Jace
//	 card from your hand.)
//	 Counter target spell. Empower Jace 1."
//
// ADR 0139 proof card. "Behold a Jace" is the ADR 0100 either/or cost
// with the Jace subtype (CR 701.4a): a Jace card in hand, or any Jace
// permanent you control — a Jace token from an earlier Empower Jace
// included, which is what makes the set's spells feed each other.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "fc3353e9-fe24-4ca7-bff8-9da767f2903a",
		Name:           "Countersculpt",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.BeholdOrPay("a", "Jace", "{1}"),
		Targets:        effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) > 0 {
				if err := (effects.CounterTarget{StackID: item.Targets[0].ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.EmpowerJace{N: 1}.Apply(ctx)
		},
	})
}
