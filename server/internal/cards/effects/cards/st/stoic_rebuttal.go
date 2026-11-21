package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stoic Rebuttal — Instant {1}{U}{U} (Scars of Mirrodin, #5577):
//
//	"Metalcraft — This spell costs {1} less to cast if you control three
//	 or more artifacts.
//	 Counter target spell."
//
// Academy Journeymage's self cost reduction (CR 601.2f) on the metalcraft
// count, taken as the spell is cast, then Counterspell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6dd5782d-2d1a-4d6b-8381-e15d0d5be24f",
		Name:         "Stoic Rebuttal",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Metalcraft — This spell costs {1} less to cast if you control three or more artifacts.", effects.SomB1MetalcraftCost()),
		},
		Targets:   effects.TargetSpell("target spell"),
		OnResolve: effects.SomB1CounterTheTargetSpells,
	})
}
