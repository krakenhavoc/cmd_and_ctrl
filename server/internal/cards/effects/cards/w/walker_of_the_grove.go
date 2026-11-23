package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Walker of the Grove — Creature — Elemental {6}{G}{G}, 7/7 (Morningtide,
// #5592):
//
//	"When this creature leaves the battlefield, create a 4/4 green
//	 Elemental creature token.
//	 Evoke {4}{G} (You may cast this spell for its evoke cost. If you do,
//	 it's sacrificed when it enters.)"
//
// Reveillark's leave trigger (any exit) making one token from the shared
// table. Evoke is the alternative cost (CR 702.74a): for {4}{G} the Walker
// enters, is sacrificed, and leaves the 4/4 behind.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "02d69413-e040-46b1-8f58-69bde666329d",
		Name:             "Walker of the Grove",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Evoke("{4}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisLeaves("Walker of the Grove — create a 4/4 green Elemental token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("4/4 green Elemental"), N: 1})),
		},
	})
}
