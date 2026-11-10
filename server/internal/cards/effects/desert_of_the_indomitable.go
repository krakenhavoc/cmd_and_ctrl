package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Desert of the Indomitable — Land — Desert:
//
//	"This land enters tapped.
//	 {T}: Add {G}.
//	 Cycling {1}{G} ({1}{G}, Discard this card: Draw a card.)"
//
// Barren Moor's shape with a different colour and cycling cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "852d4dc3-404d-4565-99e9-1eac8f6eca5e",
		Name:          "Desert of the Indomitable",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		Activated:     []ActivatedAbility{Cycling("{1}{G}")},
		ManaAbilities: []ManaAbility{cmdr2B01TapForColor("G")},
	})
}
