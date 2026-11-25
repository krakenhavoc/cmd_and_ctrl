package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jetting Glasskite — {4}{U}{U} Creature — Spirit 4/4 (Eternal Masters, S82
// OLD, #5542):
//
//	"Flying
//	 Whenever this creature becomes the target of a spell or ability for the
//	 first time each turn, counter that spell or ability."
//
// Kira, Great Glass-Spinner's grant on the Glasskite itself: the trigger counts
// per object and per turn, for a spell or ability of any controller, and counters
// the stack object named by the event (a spell or an ability alike, CR 701.6).
// The trigger counts from the moment it is put on the stack, so a first
// targeting that was responded to still used the turn's first.
//
// No simplification.
func init() {
	const label = "Jetting Glasskite — counter that spell or ability"
	effects.Register(effects.Spec{
		OracleID:        "65b61373-835a-4b22-90d1-d019446e98f7",
		Name:            "Jetting Glasskite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{effects.CounterTheFirstTargetingEachTurn(label)},
	})
}
