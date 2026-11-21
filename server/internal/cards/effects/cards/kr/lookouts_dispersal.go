package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lookout's Dispersal — Instant {2}{U} (Jumpstart 2022, #5948):
//
//	"This spell costs {1} less to cast if you control a Pirate.
//	 Counter target spell unless its controller pays {4}."
//
// Academy Journeymage's conditional self cost reduction (the Pirate is read
// through the layers as the spell is cast) on a pay-unless counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e2e44d46-9df4-46f8-acf7-5581b2257620",
		Name:         "Lookout's Dispersal",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(1, "This spell costs {1} less to cast if you control a Pirate.", effects.YouControlA(effects.And(effects.Creature(), effects.HasSubtype("Pirate")))),
		},
		Targets: effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.CounterTheTargetSpellUnlessPaid("Lookout's Dispersal", "{4}")(ctx.Game, item)
		},
	})
}
