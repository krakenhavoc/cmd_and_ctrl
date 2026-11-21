package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "2b1bec38-5aaf-4ca0-9a04-fb4367c59c1e",
		Name:         "Run Over",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfItTargets(1, "This spell costs {1} less to cast if it targets a Mount or Vehicle you control.",
				effects.And(effects.Or(effects.HasSubtype("Mount"), effects.HasSubtype("Vehicle")), effects.YouControl())),
		},
		OnResolve: effects.OneSidedBite,
	})
}
