package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Victim of Night — Instant {B}{B} (Innistrad, #5572):
//
//	"Destroy target non-Vampire, non-Werewolf, non-Zombie creature."
//
// A creature with none of the three creature types (AnySubtype, read
// through changed text, CR 612.2), destroyed by destroyTheTargetPermanent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d92b9f63-17b6-4ed2-b6ae-083a8161863e",
		Name:         "Victim of Night",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: TargetCreature("target non-Vampire, non-Werewolf, non-Zombie creature",
			Not(AnySubtype("Vampire", "Werewolf", "Zombie"))),
		OnResolve: destroyTheTargetPermanent,
	})
}
