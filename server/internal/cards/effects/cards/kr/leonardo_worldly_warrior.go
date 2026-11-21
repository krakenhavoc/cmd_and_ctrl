package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leonardo, Worldly Warrior — Legendary Creature — Mutant Ninja Turtle
// {7}{W}, 5/5 (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"Affinity for creatures (This spell costs {1} less to cast for each
//	 creature you control.)
//	 Double strike"
//
// Witherbloom's self cost modifier (CR 702.41a): {1} less per creature
// its caster controls as the cost is determined, generic only, so the
// {W} is always paid. The mana value stays 8 (the card's ruling). Double
// strike rides PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "94dcbcc0-6941-4e8c-8741-4b47ae231eba",
		Name:            "Leonardo, Worldly Warrior",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for creatures", effects.Creature()),
		},
	})
}
