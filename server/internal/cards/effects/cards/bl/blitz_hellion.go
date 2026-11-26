package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blitz Hellion — Creature — Hellion {3}{R}{G}, 7/7 (Alara Reborn, #5586):
//
//	"Trample, haste
//	 At the beginning of the end step, this creature's owner shuffles it into
//	 their library."
//
// Every end step (any player's). The owner, not the controller, shuffles it in,
// through the ordinary tuck (a commander's owner is offered the command zone,
// CR 903.9), then shuffles; a Hellion that has left does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7ef3f240-05a5-4fc9-a977-a729010a53c2",
		Name:            "Blitz Hellion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.AtEachStep(game.StepEnd, "Blitz Hellion — its owner shuffles it into their library", effects.S82OldB504ShuffleThisIntoLibrary),
		},
	})
}
