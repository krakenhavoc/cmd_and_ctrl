package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "6dd5782d-2d1a-4d6b-8381-e15d0d5be24f",
		Name:         "Stoic Rebuttal",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLess(1, "Metalcraft — This spell costs {1} less to cast if you control three or more artifacts.", somB1MetalcraftCost()),
		},
		Targets:   TargetSpell("target spell"),
		OnResolve: somB1CounterTheTargetSpells,
	})
}
