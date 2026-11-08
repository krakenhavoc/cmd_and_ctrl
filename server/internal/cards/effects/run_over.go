package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Run Over — Instant {1}{G} (Aetherdrift, #4464):
//
//	"This spell costs {1} less to cast if it targets a Mount or Vehicle
//	 you control.
//	 Target creature you control deals damage equal to its power to
//	 target creature an opponent controls."
//
// Bite Down's two clauses and body (oneSidedBite), with the second clause
// limited to a creature an opponent controls. The power is read as the
// spell resolves (CR 608.2h). The reduction is CostsLessIfItTargets over
// a Mount or Vehicle its caster controls, judged against the announced
// targets when the cost is determined (CR 601.2f): a crewed Vehicle or a
// Mount picked as the biter qualifies, an opponent's Vehicle does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2b1bec38-5aaf-4ca0-9a04-fb4367c59c1e",
		Name:         "Run Over",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(1, "This spell costs {1} less to cast if it targets a Mount or Vehicle you control.",
				And(Or(HasSubtype("Mount"), HasSubtype("Vehicle")), YouControl())),
		},
		OnResolve: oneSidedBite,
	})
}
