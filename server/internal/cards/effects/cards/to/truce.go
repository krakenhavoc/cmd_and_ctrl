package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Truce — Instant {2}{W} (Homelands, S82 OLD b826, #5635):
//
//	"Each player may draw up to two cards. For each card less than two a
//	 player draws this way, that player gains 2 life."
//
// Temporary Truce's sentence at instant speed: each player, the active player
// first and then in turn order, picks 0 to two (a draw limit caps it, CR
// 121.2b), draws that many and gains 2 life for each card short of two,
// counted from the draws that really arrived (OldB826DrawUpToTwo).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fe4aa72b-f176-461a-aca0-1519721609b2",
		Name:         "Truce",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.OldB826DrawUpToTwo("truce", "Truce"),
	})
}
