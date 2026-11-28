package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rain of Salt — Sorcery {4}{R}{R} (Urza's Saga, S82 OLD batch 729, #5624):
//
//	"Destroy two target lands."
//
// Exactly two target lands, any player's, chosen as the spell is cast; each is
// rechecked as it resolves and one that has gone or stopped being a land is
// skipped (CR 608.2b). The ones still legal are destroyed together
// (CR 700.4), as Curtains' Call's two creatures are.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1219e330-01ac-405a-b75a-dd4298598167",
		Name:         "Rain of Salt",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("two target lands", effects.Land()).WithCount(2, 2),
		OnResolve:    effects.DestroyEachLegalTarget,
	})
}
