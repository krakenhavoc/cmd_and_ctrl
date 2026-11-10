package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Waxen Shapethief — Creature — Shapeshifter {3}{U}, 0/0 (Aetherdrift,
// #4464):
//
//	"Flash
//	 You may have this creature enter as a copy of an artifact or creature
//	 you control.
//	 Cycling {2}"
//
// Chameleon, Master of Disguise's copy replacement (CR 707.2, 614.12) over
// "an artifact or creature you control": the candidates are the permanents
// of the entering creature's controller that are artifacts or creatures,
// itself excluded, checked when the prompt is built and again on the
// answer, and picked without targeting. A copy of an artifact that is not a
// creature stays a noncreature artifact, and one that is a Vehicle is a
// Vehicle. The copy has the copied object's copiable values, so its flash
// and its cycling come from what it copied, not from the Shapethief. Declined
// (or with nothing to copy) it is a 0/0 that dies as a state-based action.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6c540822-4ceb-4ebf-b777-da172ada3f1b",
		Name:            "Waxen Shapethief",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Activated:       []ActivatedAbility{Cycling("{2}")},
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf(
				"Waxen Shapethief",
				func(g *game.Game, controller uuid.UUID, self uuid.UUID) []uuid.UUID {
					return copyCandidates(g, self, func(c game.Card) bool {
						return c.Controller == controller && (c.IsCreature() || c.IsArtifact())
					})
				},
				nil,
			),
		},
	})
}
