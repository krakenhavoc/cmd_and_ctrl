package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scattered Groves — Land — Forest Plains (Marvel Super Heroes
// Commander, #4455):
//
//	"({T}: Add {G} or {W}.)
//	 This land enters tapped.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The green-white member of the cycling-land cycle: Canyon Slough's
// shape with its colours.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "3c87ea85-ca29-45a7-b5b2-758c62898b0a",
		Name:          "Scattered Groves",
		Completeness:  CompletenessFull,
		Activated:     []ActivatedAbility{Cycling("{2}")},
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("G", "W")},
	})
}
