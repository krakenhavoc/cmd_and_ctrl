package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Voq, Torchbearer — Legendary Creature — Klingon Spy {3}{B}, 4/3:
//
//	"Menace
//	 ChoH'a' — You may have Voq enter as a copy of a creature card in a
//	 graveyard, except he's a Klingon Spy in addition to his other types
//	 and has menace."
//
// An enters-as-a-copy of a graveyard card (#4797, CR 707.2, CR 614.1c).
// The except clause is part of the copy effect (CR 707.9b), so the copy
// is a Klingon Spy with menace whatever it copies. Declined, Voq enters
// as his printed self, which already has menace.
func init() {
	Register(Spec{
		OracleID:        "1b8be000-e2fd-4157-ace2-7b8603c03fba",
		Name:            "Voq, Torchbearer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf("Voq, Torchbearer", anyCreatureCardInAGraveyard,
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.AddSubtype("Klingon")
					v.AddSubtype("Spy")
					v.AddKeyword("menace")
				}),
		},
	})
}
