package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shimmering Glasskite — Creature — Spirit {3}{U}, 2/3 (Betrayers of
// Kamigawa, #5602):
//
//	"Flying
//	 Whenever this creature becomes the target of a spell or ability for the
//	 first time each turn, counter that spell or ability."
//
// Jetting Glasskite's trigger: it counts per object and per turn, for a spell
// or ability of any controller, and counters the stack object named by the
// event (WhenFirstTargetedEachTurnCounterIt, CR 701.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0a661f7a-5d56-4d73-bab7-22bfc5cef22e",
		Name:            "Shimmering Glasskite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{effects.WhenFirstTargetedEachTurnCounterIt("Shimmering Glasskite — counter that spell or ability")},
	})
}
