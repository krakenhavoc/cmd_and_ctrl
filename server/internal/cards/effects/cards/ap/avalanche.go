package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Avalanche — Sorcery {X}{2}{R}{R} (Ice Age, S82 OLD, #5636):
//
//	"Destroy X target snow lands."
//
// Exactly X targets (CR 601.2c): the count is the announced X, so the spell can't be cast for more than the snow
// lands there are. Mishra's Helix's X-target clause; the lands are destroyed together.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a349e526-882c-454f-ba79-ad17340f8153",
		Name:         "Avalanche",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.S82OldB844AvalancheTargets(),
		OnResolve:    effects.DestroyEachLegalTarget,
	})
}
