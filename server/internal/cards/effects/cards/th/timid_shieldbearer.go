package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Timid Shieldbearer — Creature — Kithkin Soldier {1}{W}, 2/2 (Lorwyn
// Eclipsed, #4458):
//
//	"{4}{W}: Creatures you control get +1/+1 until end of turn."
//
// An anthem for the turn: the set is the creatures its controller controls
// as the ability resolves, fixed then (CR 611.2c), so one that enters later
// in the turn gets nothing. It can be activated any number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8bef3511-ceeb-4d57-972a-ab363ecacd5b",
		Name:         "Timid Shieldbearer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{W}: Creatures you control get +1/+1 until end of turn.",
			Cost:    effects.ManaCost("{4}{W}"),
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1, Toughness: 1}, Answers: game.AnswerPump},
			Effect: effects.Do(effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1,
				Label: "Timid Shieldbearer — +1/+1"}),
		}},
	})
}
