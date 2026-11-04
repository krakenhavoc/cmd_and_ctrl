package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Body Double — Creature — Shapeshifter {4}{U}, 0/0:
//
//	"You may have this creature enter as a copy of any creature card in a
//	 graveyard."
//
// Clone with a graveyard for a battlefield (#4797, CR 707.2, CR 614.1c):
// the copy_target prompt offers every creature card in every graveyard,
// and the copy takes the card's printed values. The card stays where it
// is. Declining, or a pick that left the graveyard before the answer,
// leaves a 0/0 that dies to CR 704.5f.
func init() {
	Register(Spec{
		OracleID:     "895527a6-e3b2-454d-a3be-c0036630ec73",
		Name:         "Body Double",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf("Body Double", anyCreatureCardInAGraveyard, nil),
		},
	})
}
