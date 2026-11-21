package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spite // Malice — split card (Invasion, #5616):
//
//	Spite — Instant {3}{U}: "Counter target noncreature spell."
//	Malice — Instant {3}{B}: "Destroy target nonblack creature. It
//	         can't be regenerated."
//
// Each half is an ADR 0034 face (ADR 0103): Spite under the bare oracle
// ID, Malice under "#1". Only one half can be cast (CR 709.3; the card
// has no fuse). Spite is Negate's clause and the shared
// counterTheTargetSpell body; Malice is Doom Blade's clause with
// Terminate's destroyTheTargetPermanentNoRegen, so a regeneration shield
// does not save the creature (CR 701.19c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b829de81-04cd-498e-9f31-074f23e2a488",
		Name:         "Spite",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target noncreature spell", effects.Noncreature()),
		OnResolve:    effects.CounterTheTargetSpell,
	})
	effects.Register(effects.Spec{
		OracleID:     "b829de81-04cd-498e-9f31-074f23e2a488#1",
		Name:         "Malice",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target nonblack creature", effects.NonBlack()),
		OnResolve:    effects.DestroyTheTargetPermanentNoRegen,
	})
}
