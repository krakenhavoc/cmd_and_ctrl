package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wall of Shards — Snow Creature — Wall {1}{W}, 1/8 (Coldsnap, S82 OLD b558):
//
//	"Defender, flying
//	 Cumulative upkeep—An opponent gains 1 life."
//
// One life gain per age counter, each given to an opponent of the payer's
// choice (the same one twice if they like; CR 702.24a), paid together or not
// at all.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8fab68ad-169d-46d3-93c4-5bdee4eea2ce",
		Name:            "Wall of Shards",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeepPaying("Wall of Shards — cumulative upkeep: an opponent gains 1 life",
				effects.OpponentGainsLifePayment(1, 1), "", ""),
		},
	})
}
