package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jwari Shapeshifter — Creature — Shapeshifter Ally {1}{U} 0/0 (Worldwake,
// S82 OLD batch 489, #5582):
//
//	"You may have this creature enter as a copy of any Ally creature on the
//	 battlefield."
//
// A copy replacement (CR 614.1c, EntersAsCopyOf) over the Ally creatures on
// the battlefield, read through the layers so a changeling is one. The
// choice is optional; declined, it is a 0/0 and dies (CR 704.5f). The Ally
// word is read through the card's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b0c02eb-4d9e-4fea-bad6-9ef5d7bb5295",
		Name:         "Jwari Shapeshifter",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf("Jwari Shapeshifter", func(g *game.Game, _ uuid.UUID, self uuid.UUID) []uuid.UUID {
				ally := "Ally"
				if me, ok := g.LookupCardForEffect(self); ok {
					ally = me.TextSubtype("Ally")
				}
				return effects.CopyCandidates(g, self, func(c game.Card) bool {
					return c.IsCreature() && c.HasSubtype(ally)
				})
			}, nil),
		},
	})
}
