package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dissipate — Instant {1}{U}{U} (Innistrad: Midnight Hunt Commander
// reprint, S82):
//
//	"Counter target spell. If that spell is countered this way, exile it
//	 instead of putting it into its owner's graveyard."
//
// Assert Authority's resolution without the affinity: CounterTarget sends
// the spell to exile, and a spell that can't be countered stays on the
// stack and is not exiled either.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b2e9aa9-6f91-41de-84b6-e9a89be06a53",
		Name:         "Dissipate",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetSpell("target spell"),
		OnResolve:    counterTargetSpellAndExileIt,
	})
}
