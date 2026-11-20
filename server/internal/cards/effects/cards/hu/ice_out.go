package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ice Out — Instant {1}{U}{U} (Wilds of Eldraine, #4471):
//
//	"Bargain (You may sacrifice an artifact, enchantment, or token as you
//	 cast this spell.)
//	 This spell costs {1} less to cast if it's bargained.
//	 Counter target spell."
//
// Bargain is the shared optional cost (g2Bargain), announced with the
// cast (CR 601.2b); the discount reads that announcement as the total
// cost is fixed (CostsLessIfBargained, CR 601.2f), so a bargained Ice
// Out costs {U}{U} and its mana value stays 3 (CR 202.3). The body is
// Cancel's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "fed3bc35-4269-48f0-9133-a595ce49c2d1",
		Name:          "Ice Out",
		Completeness:  effects.CompletenessFull,
		Targets:       effects.TargetSpell("target spell"),
		OptionalCosts: []game.AdditionalCost{effects.G2Bargain()},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfBargained(1, "This spell costs {1} less to cast if it's bargained."),
		},
		OnResolve: effects.CounterTheTargetSpell,
	})
}
