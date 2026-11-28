package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyshroud Blessing — Instant {1}{G} (Planeshift, S82 OLD #5615):
//
//	"All lands gain shroud until end of turn. (They can't be the targets of
//	 spells or abilities.)
//	 Draw a card."
//
// Every land on the battlefield as the spell resolves gains shroud until end
// of turn (CR 611.2c: the set is fixed then), then the draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ffcedb73-538a-45ea-8564-7b92624537e0",
		Name:         "Skyshroud Blessing",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerProtect, Draws: 1},
		OnResolve: effects.ResolveSteps(
			effects.GrantKeywordUntilEOT{Match: effects.Land(), Keywords: []string{"shroud"}, Label: "Skyshroud Blessing — all lands gain shroud"},
			effects.DrawCards{N: 1},
		),
	})
}
