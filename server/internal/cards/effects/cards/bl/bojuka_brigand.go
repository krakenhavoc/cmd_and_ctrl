package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bojuka Brigand — Creature — Human Warrior Ally {1}{B}, 1/1 (Worldwake, #5582):
//
//	"This creature can't block.
//	 Whenever this creature or another Ally you control enters, you may
//	 put a +1/+1 counter on this creature."
//
// The Ally trigger is South Pole Voyager's: this creature's own entry,
// or another Ally entering under its controller's control
// (wwkThisOrAnotherAllyEntered). The +1/+1 counter is a CR 603.5 "you
// may" asked as the trigger resolves (wwkAllyMayCounter).
// Its "can't block" is Bloodghast's static restriction.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1d82070-3198-4ead-9431-d90661e29741",
		Name:         "Bojuka Brigand",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Triggered:    []game.TriggeredAbility{effects.WwkAllyMayCounter("Bojuka Brigand")},
	})
}
