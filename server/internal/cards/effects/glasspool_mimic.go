package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glasspool Mimic // Glasspool Shore — Modal DFC. Front: Creature —
// Shapeshifter Rogue {2}{U}, 0/0:
//
//	"You may have this creature enter as a copy of a creature you
//	 control, except it's a Shapeshifter Rogue in addition to its other
//	 types."
//
// Back: Land, "This land enters tapped. {T}: Add {U}." (the back face
// is in mdfc_lands.go).
//
// Clone's shape (ADR 0043) narrowed to "a creature you control": the
// candidates are the controller's own creatures, not targeted (CR
// 115.10a), so hexproof does not matter, and the copy is optional. With
// nothing to copy, or if the player declines, it enters as the printed
// 0/0 and dies (CR 704.5f). The "except" clause adds both Shapeshifter
// and Rogue to the copied subtypes, and the oracle ID rides the copy, so
// the copied creature's triggers, statics and abilities come with it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c178953c-3888-4edd-9d0c-265bd82b1d24",
		Name:         "Glasspool Mimic",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf(
				"Glasspool Mimic",
				func(g *game.Game, controller uuid.UUID, self uuid.UUID) []uuid.UUID {
					return copyCandidates(g, self, func(c game.Card) bool {
						return c.IsCreature() && c.Controller == controller
					})
				},
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.AddSubtype("Shapeshifter")
					v.AddSubtype("Rogue")
				}),
		},
	})
}
