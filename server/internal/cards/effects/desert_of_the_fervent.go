package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Desert of the Fervent — Land — Desert:
//
//	"This land enters tapped.
//	 {T}: Add {R}.
//	 Cycling {1}{R} ({1}{R}, Discard this card: Draw a card.)"
//
// Barren Moor's shape with a different colour and cycling cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "311f38a6-f68f-4d30-bc4e-62339f1e0d88",
		Name:          "Desert of the Fervent",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		Activated:     []ActivatedAbility{Cycling("{1}{R}")},
		ManaAbilities: []ManaAbility{cmdr2B01TapForColor("R")},
	})
}
