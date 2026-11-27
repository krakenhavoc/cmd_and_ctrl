package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Maga, Traitor to Mortals — Legendary Creature — Human Wizard {X}{B}{B}{B}, 0/0
// (Saviors of Kamigawa, S82 OLD b593):
//
//	"Maga enters with X +1/+1 counters on it.
//	 When Maga enters, target player loses life equal to the number of +1/+1
//	 counters on it."
//
// The counters are the printed CR 614.1c entry clause (XCounters). The trigger counts
// them as it resolves, from last-known information if Maga has already gone (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "1fb1afc0-a3bf-4368-9ca0-47dd8c4c3c60",
		Name:                       "Maga, Traitor to Mortals",
		Completeness:               effects.CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterPlusOne)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Maga — target player loses life equal to the number of +1/+1 counters on it", effects.MagaDrain),
				effects.TargetPlayer("target player")),
		},
	})
}
