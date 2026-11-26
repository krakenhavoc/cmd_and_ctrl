package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Polymorphous Rush — Instant {2}{U} (Journey into Nyx):
//
//	"Strive — This spell costs {1}{U} more to cast for each target beyond the first.
//	 Choose a creature on the battlefield. Any number of target creatures you
//	 control each become a copy of that creature until end of turn."
//
// The choice is made as the spell resolves, from every creature on the
// battlefield (it is not a target, so hexproof and shroud do not stop it).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3098626a-b786-4f54-83df-51833729c1ec",
		Name:         "Polymorphous Rush",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetCreature("any number of target creatures you control", effects.YouControl()).WithCount(0, 0),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsMorePerTargetBeyondFirst("{1}{U}", "Strive — This spell costs {1}{U} more to cast for each target beyond the first."),
		},
		OnResolve: effects.S82OldB366ChooseCreatureThenCopy,
	})
}
