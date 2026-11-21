package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Essence of Antiquity — Artifact Creature — Golem {3}{W}{W}, 1/10
// (Murders at Karlov Manor, #4469):
//
//	"Disguise {2}{W} (You may cast this card face down for {3} as a 2/2
//	 creature with ward {2}. Turn it face up any time for its disguise
//	 cost.)
//	 When this creature is turned face up, creatures you control gain
//	 hexproof until end of turn. Untap them."
//
// The face-up trigger (CR 708.8) is Glaring Spotlight's team grant: the
// creatures its controller controls as it resolves gain hexproof until
// end of turn (CR 611.2c), then each creature they control is untapped
// (UntapAllCreaturesYouControl), the same set.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "26a7f92d-6914-466d-b7c7-9ac121dd5118",
		Name:             "Essence of Antiquity",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Disguise("{2}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisIsTurnedFaceUp("Essence of Antiquity — creatures you control gain hexproof until end of turn; untap them",
				effects.Do(
					effects.GrantKeywordUntilEOT{
						Match:    effects.And(effects.Creature(), effects.YouControl()),
						Keywords: []string{"hexproof"},
						Label:    "Essence of Antiquity — hexproof",
					},
					effects.UntapAllCreaturesYouControl{},
				)),
		},
	})
}
