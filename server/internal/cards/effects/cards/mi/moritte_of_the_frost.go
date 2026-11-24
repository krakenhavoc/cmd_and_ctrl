package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Moritte of the Frost — Legendary Snow Creature — Shapeshifter {2}{G}{U}{U},
// 0/0:
//
//	"Changeling (This card is every creature type.)
//	 You may have Moritte enter as a copy of a permanent you control, except
//	 it's legendary and snow in addition to its other types and, if it's a
//	 creature, it enters with two additional +1/+1 counters on it and has
//	 changeling."
//
// Spark Double's enter-as-a-copy over any permanent its controller controls.
// The exceptions are on the copiable values (CR 707.9a), so a copy of a copy
// keeps them: Legendary and Snow are added to its supertypes, and when it is a
// creature it also has changeling and the two counters arrive with the entry,
// before any enters trigger. With nothing to copy, or the copy declined, it is
// its own changeling 0/0 and dies as a state-based action (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "78699161-fc14-4a44-8a15-0f7c08be0343",
		Name:            "Moritte of the Frost",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"changeling"},
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf(
				"Moritte of the Frost",
				func(g *game.Game, controller uuid.UUID, self uuid.UUID) []uuid.UUID {
					return effects.CopyCandidates(g, self, func(c game.Card) bool { return c.Controller == controller })
				},
				func(ev *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.AddSupertype("Legendary")
					v.AddSupertype("Snow")
					if v.HasCardType("Creature") {
						v.AddKeyword("changeling")
						ev.AddCounterAtETB(game.CounterPlusOne, 2)
					}
				},
			),
		},
	})
}
