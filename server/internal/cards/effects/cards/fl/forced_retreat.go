package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forced Retreat — Sorcery {2}{U}
// (Portal Three Kingdoms, #5622):
//
//	"Put target creature on top of its owner's library."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ccf1336c-2ea5-4582-84a5-0b29b3cca6c2",
		Name:         "Forced Retreat",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.GapB79PutTargetOnTopOfItsOwnersLibrary,
	})
}
