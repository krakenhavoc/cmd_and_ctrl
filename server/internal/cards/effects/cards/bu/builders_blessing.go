package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Builder's Blessing — {3}{W} Enchantment: "Untapped creatures you control get
// +0/+2."
//
// A layer 7c anthem whose subject is re-read every pass, so a creature that
// taps loses the bonus and one that untaps gets it back.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "33a64de1-dc78-43d3-b89e-b4b13049e174",
		Name:         "Builder's Blessing",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.B16Anthem(effects.UntappedCreatureYouControl, 0, 2)},
	})
}
