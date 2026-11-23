package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Time Beetle — Creature — Alien Insect {1}{U}, 1/1 (Doctor Who, #5948):
//
//	"Skulk (This creature can't be blocked by creatures with greater power.)
//	 Parallel Universe — Whenever this creature deals combat damage to a
//	 player, time travel."
//
// Skulk is a printed keyword. The trigger is TimeTravel (CR 701.56a): for each
// suspended card the controller owns and each permanent they control with a
// time counter on it, they may add or remove one counter. "Parallel
// Universe" is an ability word (CR 207.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "247a0c13-c611-4c00-bedf-ecf7ea6bb956",
		Name:            "Time Beetle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"skulk"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Time Beetle — time travel", effects.Do(effects.TimeTravel{})),
		},
	})
}
