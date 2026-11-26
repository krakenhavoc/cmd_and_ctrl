package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Essence of the Wild — Creature — Avatar {3}{G}{G}{G}, 6/6 (Innistrad,
// S82 OLD batch 450, #5572):
//
//	"Creatures you control enter as a copy of this creature."
//
// The static enters-as replacement of #6262 (CR 614.1c, 707.2): every
// creature entering under its controller's control, token or card, enters
// as a copy of this creature, which has the replacement itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d105f83-583d-438b-8f26-b39e818ed295",
		Name:         "Essence of the Wild",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.CreaturesYouControlEnterAsCopyOfThis("Creatures you control enter as a copy of this creature."),
		},
	})
}
