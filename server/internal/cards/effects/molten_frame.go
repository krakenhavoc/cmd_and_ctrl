package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Molten Frame — Instant {1}{R} (Conflux, #5587):
//
//	"Destroy target artifact creature.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Destroy through the shared destroyTheTargetPermanent body (CR 701.8a);
// a creature that stopped being an artifact in response is no longer a
// legal target (CR 608.2b). With the engine's cycling row (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7c72df45-c3bd-4fb9-8365-5a5639566742",
		Name:         "Molten Frame",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target artifact creature", Artifact()),
		OnResolve:    destroyTheTargetPermanent,
		Activated:    []ActivatedAbility{Cycling("{2}")},
	})
}
