package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Synth Infiltrator — Artifact Creature — Synth {3}{U}{U}, 0/0:
//
//	"Improvise ... You may have this creature enter as a copy of any
//	 creature on the battlefield, except it's a Synth artifact creature
//	 in addition to its other types."
//
// Improvise is Spec.TapCost; the copy is Clever Impersonator's
// EntersAsCopyOf restricted to creatures, with the Synth/artifact
// exception applied to the copiable values. No simplification.
func init() {
	Register(Spec{
		OracleID:     "3d3553df-0036-47d9-b0e3-4a3261a0f73c",
		Name:         "Synth Infiltrator",
		Completeness: CompletenessFull,
		TapCost:      Improvise(),
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf(
				"Synth Infiltrator",
				func(g *game.Game, _ uuid.UUID, self uuid.UUID) []uuid.UUID {
					return copyCandidates(g, self, func(c game.Card) bool { return c.IsCreature() })
				},
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.AddCardType("Artifact")
					v.AddSubtype("Synth")
				},
			),
		},
	})
}
