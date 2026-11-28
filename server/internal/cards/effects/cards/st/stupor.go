package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stupor — Sorcery {2}{B} (Mirage, S82 OLD b799, #5633):
//
//	"Target opponent discards a card at random, then discards a card."
//
// The random discard first, as a keyed discard whose tail asks the same
// opponent to discard a card of their choice once it has settled (CR
// 608.2c). A hand emptied by the first discards nothing more.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c19e3560-d304-4d91-aa2a-463d7dbb35bb",
		Name:         "Stupor",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 2}),
		OnResolve:    effects.OldB799Stupor,
	})
}
