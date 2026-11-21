package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hulking Metamorph — Artifact Creature — Shapeshifter {9}, 7/7 (The
// Brothers' War, #5656): "Prototype {2}{U}{U} — 3/3", "You may have this
// creature enter as a copy of an artifact or creature you control, except
// it's an artifact creature in addition to its other types, and its power
// and toughness are equal to this creature's power and toughness."
//
// Prototype is the card's second face (CR 718, game/prototype.go), built
// by the deck importer from the inset; nothing here spells it out. The
// copy replacement is Phyrexian Metamorph's (EntersAsCopyOf, CR 614.1c,
// CR 707.9) over the artifacts and creatures its controller controls. The
// except clause adds the artifact and creature types and sets the power and
// toughness to those of the Metamorph being cast, as printed or prototyped:
// 3/3 for a prototyped one, 7/7 otherwise (CR 707.9b). Not targeting, so a
// hexproof permanent can be copied. Declined, or with nothing to copy, it
// enters as itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b2b09dc3-8239-4f3b-91d2-e7cbbdc062f7",
		Name:         "Hulking Metamorph",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf(
				"Hulking Metamorph",
				func(g *game.Game, controller uuid.UUID, self uuid.UUID) []uuid.UUID {
					return effects.CopyCandidates(g, self, func(c game.Card) bool {
						return c.Controller == controller && (c.IsArtifact() || c.IsCreature())
					})
				},
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, source *game.Card) {
					v.AddCardType("Artifact")
					v.AddCardType("Creature")
					if source != nil {
						eff := source.Effective()
						v.SetPowerToughness(eff.Power, eff.Toughness)
					}
				},
			),
		},
	})
}
