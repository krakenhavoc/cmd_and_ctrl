package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Powder Ganger — Creature — Human Rogue {2}{R}, 2/2 (Fallout, #5469):
//
//	"Squad {2} (As an additional cost to cast this spell, you may pay
//	 {2} any number of times. When this creature enters, create that
//	 many tokens that are copies of it.)
//	 When this creature enters, destroy up to one target artifact."
//
// Imperial Hullbreachers' squad (SquadMana, Squad: CR 702.157a, the
// engine's cap of 20 payments on one announcement, multikicker's
// reason) and an enters trigger with an "up to one target artifact"
// clause. Each token copy has the enters trigger too, so it destroys up
// to one more artifact as it enters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "b34507d3-55ba-4629-9581-7983d0025e89",
		Name:          "Powder Ganger",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.SquadMana("{2}", 20)},
		Triggered: []game.TriggeredAbility{
			effects.Squad("Powder Ganger"),
			effects.Targeting(effects.WhenThisEnters("Powder Ganger — destroy up to one target artifact", effects.DestroyFirstLegalCardTarget),
				effects.TargetPermanent("up to one target artifact", effects.Artifact()).WithCount(0, 1)),
		},
	})
}
