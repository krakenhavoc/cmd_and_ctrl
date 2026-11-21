package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Death-Greeter's Champion — Creature — Human Warrior {2}{R}, 2/1 (March of the Machine Commander):
//
//	"Dash {3}{R}
//	 Backup 1 (When this creature enters, put a +1/+1 counter on target creature.
//	 If that's another creature, it gains the following ability until end of turn.)
//	 Double strike"
//
// Backup (CR 702.165) is the shared Backup constructor (soc_b6_helpers.go):
// the counters go on the chosen creature, and another creature also gains
// double strike until end of turn; the card has it itself, declared in PrintedKeywords. Dash {3}{R} is the shared Dash constructor (freerunning.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "81abb451-5ee0-4f90-bacc-a8ff69726a27",
		Name:             "Death-Greeter's Champion",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"double strike"},
		AlternativeCosts: []game.AlternativeCost{effects.Dash("{3}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.Backup("Death-Greeter's Champion", 1, "", "double strike"),
		},
	})
}
