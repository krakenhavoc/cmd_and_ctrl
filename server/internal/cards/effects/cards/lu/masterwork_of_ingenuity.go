package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Masterwork of Ingenuity — Artifact — Equipment {1}:
//
//	"You may have this Equipment enter as a copy of any Equipment on the
//	 battlefield."
//
// Clone's replacement (CR 614.1c, CR 707.2) with the candidates narrowed to
// Equipment. The choice is made as it enters; a table with no other
// Equipment, or a player who declines, keeps the plain {1} Equipment with no
// abilities. Copying an Equipment does not copy what is attached to it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "35d458fb-dfdb-41e4-92e3-f0fd392e1c69",
		Name:         "Masterwork of Ingenuity",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf(
				"Masterwork of Ingenuity",
				func(g *game.Game, _ uuid.UUID, self uuid.UUID) []uuid.UUID {
					return effects.CopyCandidates(g, self, func(c game.Card) bool { return c.HasSubtype("Equipment") })
				},
				nil,
			),
		},
	})
}
