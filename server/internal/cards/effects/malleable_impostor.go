package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Malleable Impostor — Creature — Faerie Shapeshifter {3}{U}, 0/0
// (Wilds of Eldraine Commander, #5368):
//
//	"Flash
//	 Flying
//	 You may have this creature enter as a copy of a creature an opponent
//	 controls, except it's a Faerie Shapeshifter in addition to its other
//	 types and it has flying."
//
// An "enters as a copy" replacement (EntersAsCopyOf) over the creatures
// the player's opponents control. The except clause (CR 707.9b) adds the two
// creature types and flying, all part of the copiable values. Flash is not
// copied, so the Impostor has it only as a spell. Declined, or with no
// opposing creature, it enters as its printed 0/0 and dies (CR 704.5f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "acd209d1-37e9-4713-b320-ee52854790bb",
		Name:            "Malleable Impostor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf(
				"Malleable Impostor",
				func(g *game.Game, controller uuid.UUID, self uuid.UUID) []uuid.UUID {
					return copyCandidates(g, self, func(c game.Card) bool {
						return c.IsCreature() && c.Controller != controller
					})
				},
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.AddSubtype("Faerie")
					v.AddSubtype("Shapeshifter")
					v.AddKeyword("flying")
				},
			),
		},
	})
}
