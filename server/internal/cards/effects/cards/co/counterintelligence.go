package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Counterintelligence — Sorcery {2}{U}{U}
// (Portal Three Kingdoms, #5622):
//
//	"Return one or two target creatures to their owners' hands."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c4e64d1-cdcf-4e96-8bf4-ed1414f6ad62",
		Name:         "Counterintelligence",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("one or two target creatures").WithCount(1, 2),
		OnResolve:    effects.ResolveSteps(effects.Instruction(effects.NeoB20BounceTargetedPermanents)),
	})
}
