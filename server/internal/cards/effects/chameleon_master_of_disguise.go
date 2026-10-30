package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chameleon, Master of Disguise — Legendary Creature — Human
// Shapeshifter Villain {3}{U}, 2/3:
//
//	"You may have Chameleon enter as a copy of a creature you control,
//	 except his name is Chameleon, Master of Disguise.
//	 Mayhem {2}{U} (You may cast this card from your graveyard for
//	 {2}{U} if you discarded it this turn. Timing rules still apply.)"
//
// Spark Double's candidate set (a creature its controller controls) and
// Sakashima the Impostor's name exception (CR 707.9a): the copy keeps
// the copied creature's everything but its name. Not targeting, so
// hexproof does not stop it. Declined, or with no creature to copy, it
// stays the printed 2/3. Mayhem (CR 702.187, #4628).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "1ce239f2-79ad-4a30-8e8f-f985045f78dc",
		Name:             "Chameleon, Master of Disguise",
		Completeness:     CompletenessFull,
		CastableZonesIf:  []game.ConditionalCastZone{MayhemCastableZone()},
		AlternativeCosts: []game.AlternativeCost{Mayhem("{2}{U}")},
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf(
				"Chameleon, Master of Disguise",
				func(g *game.Game, controller uuid.UUID, self uuid.UUID) []uuid.UUID {
					return copyCandidates(g, self, func(c game.Card) bool {
						return c.Controller == controller && c.IsCreature()
					})
				},
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.SetName("Chameleon, Master of Disguise")
				},
			),
		},
	})
}
