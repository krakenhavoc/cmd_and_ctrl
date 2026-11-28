package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prismatic Lace — Instant {U} (Mirage, S82 OLD b799, #5633):
//
//	"Target permanent becomes the color or colors of your choice. (This effect
//	 lasts indefinitely.)"
//
// Quickchange's color-or-colors prompt (#6146), with no end: the target
// becomes exactly the chosen colors (CR 105.3, layer 5) and keeps them for as
// long as it stays on the battlefield (CR 611.2a). A target gone by
// resolution makes the spell do nothing and asks nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d4a3c9a5-3c91-4135-a19d-98eb5bcf5e00",
		Name:         "Prismatic Lace",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target permanent"),
		Purpose:      game.Purpose{Answers: game.AnswerProtect},
		OnResolve:    effects.OldB799PrismaticLace,
	})
}
