package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Saved by the Shell — {1}{G} Instant (Teenage Mutant Ninja Turtles, #4457):
//
//	"This spell costs {1} less to cast if you control a Turtle.
//	 Put a +1/+1 counter on target creature you control. It gains trample,
//	 hexproof, and indestructible until end of turn."
//
// The reduction is Wizard's Retort's, read as the cost is determined (CR
// 601.2f). The body is Snakeskin Veil's counter then grant, with three
// keywords (tmtB6CounterThenGrant).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7ef5c812-ad7d-4d6c-8ee5-68561e7e4bd7",
		Name:         "Saved by the Shell",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLess(1, "This spell costs {1} less to cast if you control a Turtle.", YouControlA(HasSubtype("Turtle"))),
		},
		Targets:   TargetCreature("target creature you control", YouControl()),
		Purpose:   ForTargets(BenefitUntilEOT(0, 0, 0, "trample", "hexproof", "indestructible")),
		OnResolve: tmtB6CounterThenGrant("Saved by the Shell — trample, hexproof, and indestructible", "trample", "hexproof", "indestructible"),
	})
}
