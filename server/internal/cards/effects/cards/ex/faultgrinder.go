package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Faultgrinder — Creature — Elemental {6}{R}, 4/4 (Commander 2011, S82 older
// sets batch 453, #5574):
//
//	"Trample
//	 When this creature enters, destroy target land.
//	 Evoke {4}{R}"
//
// Shriekmaw's evoke (sacrificed on entry, so the enters trigger and the
// sacrifice trigger go on the stack together) with a land for the target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "4aa34391-d532-48e5-a125-2619d04f27ff",
		Name:             "Faultgrinder",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"trample"},
		AlternativeCosts: []game.AlternativeCost{effects.Evoke("{4}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Faultgrinder — destroy target land", effects.DestroyChosenPermanent),
				effects.TargetPermanent("target land", effects.Land())),
		},
	})
}
