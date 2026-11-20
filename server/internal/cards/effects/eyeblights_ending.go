package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eyeblight's Ending — Kindred Instant — Elf {2}{B} (Lorwyn, #5593):
//
//	"Destroy target non-Elf creature."
//
// Doom Blade with a creature-type bar: the target's types are read through
// the layers as it is chosen and again as the spell resolves (CR 608.2b),
// so a changeling, which is an Elf (CR 702.73a), can't be targeted, and a
// creature that becomes an Elf in response makes the spell do nothing. The
// spell is a Kindred card (CR 308.1) and is cast as an instant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f3e1eef-31db-4e1a-b5a9-ac835f6a3eab",
		Name:         "Eyeblight's Ending",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target non-Elf creature", Not(HasSubtype("Elf"))),
		OnResolve:    destroyTheTargetPermanent,
	})
}
