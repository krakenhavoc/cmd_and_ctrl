package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:      "311f38a6-f68f-4d30-bc4e-62339f1e0d88",
		Name:          "Desert of the Fervent",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		Activated:     []effects.ActivatedAbility{effects.Cycling("{1}{R}")},
		ManaAbilities: []effects.ManaAbility{effects.Cmdr2B01TapForColor("R")},
	})
}
