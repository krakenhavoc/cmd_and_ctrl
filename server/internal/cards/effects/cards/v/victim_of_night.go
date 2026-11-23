package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Victim of Night — Instant {B}{B} (Innistrad, #5572):
//
//	"Destroy target non-Vampire, non-Werewolf, non-Zombie creature."
//
// A creature with none of the three creature types (AnySubtype, read
// through changed text, CR 612.2), destroyed by destroyTheTargetPermanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d92b9f63-17b6-4ed2-b6ae-083a8161863e",
		Name:         "Victim of Night",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.TargetCreature("target non-Vampire, non-Werewolf, non-Zombie creature",
			effects.Not(effects.AnySubtype("Vampire", "Werewolf", "Zombie"))),
		OnResolve: effects.DestroyTheTargetPermanent,
	})
}
