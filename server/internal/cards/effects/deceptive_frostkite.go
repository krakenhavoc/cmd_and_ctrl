package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deceptive Frostkite — Creature — Dragon {U}{U}, 1/1 (Tarkir: Dragonstorm
// Commander):
//
//	"Flying
//	 You may have this creature enter as a copy of a creature you control with
//	 power 4 or greater, except it's a Dragon in addition to its other types
//	 and it has flying."
//
// Synth Infiltrator's and Phantasmal Image's enters-as-a-copy replacement
// (CR 614.12, 707.9) with the candidates narrowed to creatures its controller
// controls whose power is 4 or greater now, read through the layers. The
// "except" adds the Dragon subtype and flying to the copiable values (CR
// 707.9a, b), so a creature that copies the Frostkite is a flying Dragon too.
// The copy is optional (CR 707.2's "you may"); declining leaves the printed
// 1/1 flying Dragon.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "96d7b761-b0b2-4ba3-98bd-3a9191ba544e",
		Name:            "Deceptive Frostkite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf(
				"Deceptive Frostkite",
				func(g *game.Game, controller uuid.UUID, self uuid.UUID) []uuid.UUID {
					return copyCandidates(g, self, func(c game.Card) bool {
						return c.IsCreature() && c.Controller == controller && b31CurrentPowerOf(g, c.InstanceID) >= 4
					})
				},
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.AddSubtype("Dragon")
					v.AddKeyword("flying")
				},
			),
		},
	})
}
