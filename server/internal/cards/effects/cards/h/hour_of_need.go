package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hour of Need — Instant {2}{U} (Journey into Nyx):
//
//	"Strive — This spell costs {1}{U} more to cast for each target beyond the first.
//	 Exile any number of target creatures. For each creature exiled this way,
//	 its controller creates a 4/4 blue Sphinx creature token with flying."
//
// Each target's controller is read before the exile (CR 608.2h); one
// token is made for each creature that actually went to exile.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2002dd67-2347-4869-98de-299074206684",
		Name:         "Hour of Need",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("any number of target creatures").WithCount(0, 0),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsMorePerTargetBeyondFirst("{1}{U}", "Strive — This spell costs {1}{U} more to cast for each target beyond the first."),
		},
		OnResolve: effects.S82OldB366ExileAndMakeSpheres,
	})
}
