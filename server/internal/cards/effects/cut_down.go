package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cut Down — Instant {B} (Dominaria United, #5487):
//
//	"Destroy target creature with total power and toughness 5 or less."
//
// The total is read through the layered view, at announce and again as the
// spell resolves (CR 608.2b), so a creature pumped past 5 in response is no
// longer a legal target and the spell does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1b30210d-10e9-4703-8967-032063bb4f26",
		Name:         "Cut Down",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // #4734: destroys a permanent
		Targets:      TargetCreature("target creature with total power and toughness 5 or less", dmuB02TotalPowerAndToughnessAtMost(5)),
		OnResolve:    ResolveSteps(Instruction(kineticOozeDestroy)),
	})
}
