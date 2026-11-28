package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Tribute — Sorcery {2}{B} (Mirage):
//
//	"As an additional cost to cast this spell, sacrifice two creatures.
//	 Destroy target artifact."
//
// The two creatures are an additional cost (CR 118.8), paid as one
// simultaneous sacrifice as the spell goes on the stack, so a caster with
// fewer than two creatures can't cast it. The artifact is destroyed if still
// legal as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "c06b993b-a36a-412a-a732-d678a4179791",
		Name:           "Phyrexian Tribute",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeNCost(2, "two creatures", effects.Creature()),
		Purpose:        game.Purpose{Answers: game.AnswerRemove},
		Targets:        effects.TargetPermanent("target artifact", effects.Artifact()),
		OnResolve:      effects.DestroyTheTargetPermanent,
	})
}
