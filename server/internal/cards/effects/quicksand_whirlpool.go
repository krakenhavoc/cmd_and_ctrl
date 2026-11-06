package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quicksand Whirlpool — Instant {5}{W} (The Lost Caverns of Ixalan,
// #4470):
//
//	"This spell costs {3} less to cast if it targets a tapped creature.
//	 Exile target creature."
//
// Ajani's Response with exile in place of destroy. The reduction reads
// the announced target (CR 601.2f, after 601.2c) through
// CostsLessIfItTargets with the tapped-creature predicate; whether the
// creature is tapped is judged as the cost is determined. The exile
// applies to whatever the target is on resolution, and the spell does
// nothing if it is no longer a legal target (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "43fb0f35-8eea-4344-9415-6ecb8aa46491",
		Name:         "Quicksand Whirlpool",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(3, "This spell costs {3} less to cast if it targets a tapped creature.", hobB3TappedCreature()),
		},
		OnResolve: ResolveSteps(ExileChosenTarget{}),
	})
}
