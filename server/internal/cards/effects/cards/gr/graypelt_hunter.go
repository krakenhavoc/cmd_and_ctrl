package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Graypelt Hunter — Creature — Human Warrior Ally {3}{G}, 2/2 (Worldwake, #5582):
//
//	"Trample
//	 Whenever this creature or another Ally you control enters, you may
//	 put a +1/+1 counter on this creature."
//
// The Ally trigger is South Pole Voyager's: this creature's own entry,
// or another Ally entering under its controller's control
// (wwkThisOrAnotherAllyEntered). The +1/+1 counter is a CR 603.5 "you
// may" asked as the trigger resolves (wwkAllyMayCounter).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b0553bf6-ecc1-4d08-9777-19827658355b",
		Name:            "Graypelt Hunter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered:       []game.TriggeredAbility{effects.WwkAllyMayCounter("Graypelt Hunter")},
	})
}
