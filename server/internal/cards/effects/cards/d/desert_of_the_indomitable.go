package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:      "852d4dc3-404d-4565-99e9-1eac8f6eca5e",
		Name:          "Desert of the Indomitable",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		Activated:     []effects.ActivatedAbility{effects.Cycling("{1}{G}")},
		ManaAbilities: []effects.ManaAbility{effects.Cmdr2B01TapForColor("G")},
	})
}
