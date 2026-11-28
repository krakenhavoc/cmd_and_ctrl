package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ether Well — Instant {3}{U} (Mirage, S82 OLD b799, #5633):
//
//	"Put target creature on top of its owner's library. If that creature is
//	 red, you may put it on the bottom of its owner's library instead."
//
// Aethertow's tuck for a creature of any color; a red creature is instead
// offered the top-or-bottom choice (the caster's, CR 401.4), read as the
// spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a2b6faca-3321-45d2-a027-ff3a8ae61f10",
		Name:         "Ether Well",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve:    effects.OldB799EtherWell,
	})
}
