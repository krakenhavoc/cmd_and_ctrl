package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Starbreach Whale — Creature — Whale {4}{U}, 3/5 (Edge of Eternities,
// #4461):
//
//	"Flying
//	 When this creature enters, surveil 2.
//	 Warp {1}{U}"
//
// A.I.M. Synthoids' entry surveil on a flier, with the shared warp cost:
// a warped Whale surveils as it enters and is exiled at the next end
// step (CR 702.185a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "076865e9-ee10-4a87-9180-4274202d37c1",
		Name:             "Starbreach Whale",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{effects.Warp("{1}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Starbreach Whale — surveil 2", effects.Do(effects.Surveil{N: 2})),
		},
	})
}
