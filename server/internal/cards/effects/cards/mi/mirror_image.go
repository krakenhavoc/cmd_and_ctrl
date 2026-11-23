package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mirror Image — Creature — Shapeshifter {2}{U}, 0/0 (Jumpstart 2022, #5948):
//
//	"You may have this creature enter as a copy of a creature you control."
//
// Clone narrowed to the controller's own creatures (EntersAsCopyOf,
// CR 707.2 + CR 614.1c). Not targeting, so hexproof does not narrow the
// choice. Declined, or with nothing to copy, it is a 0/0 and dies
// (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7275fb8c-571a-4c4c-889a-829419332e51",
		Name:         "Mirror Image",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf("Mirror Image", func(g *game.Game, controller uuid.UUID, self uuid.UUID) []uuid.UUID {
				return effects.CopyCandidates(g, self, func(c game.Card) bool {
					return c.IsCreature() && c.Controller == controller
				})
			}, nil),
		},
	})
}
