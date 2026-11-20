package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deny Existence — Instant {2}{U} (Shadows over Innistrad, #5543):
//
//	"Counter target creature spell. If that spell is countered this way,
//	 exile it instead of putting it into its owner's graveyard."
//
// Dissipate's counter-and-exile (counterTargetSpellAndExileIt) on a
// creature spell, as False Summoning targets one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b665a64b-4772-42dd-9fd2-fd8598e689ba",
		Name:         "Deny Existence",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetSpell("target creature spell", Creature()),
		OnResolve:    counterTargetSpellAndExileIt,
	})
}
