package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cliffside Lookout — Creature — Kor Scout Ally {W}, 1/1 (Battle for
// Zendikar, #5547):
//
//	"{4}{W}: Creatures you control get +1/+1 until end of turn."
//
// Timid Shieldbearer's ability: the creatures are fixed as it resolves
// (CR 611.2c), and it can be activated any number of times, the turn it
// arrives included, since its cost has no {T}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97671aac-7322-44ae-8908-d573734db3af",
		Name:         "Cliffside Lookout",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{W}: Creatures you control get +1/+1 until end of turn.",
			Cost:    effects.ManaCost("{4}{W}"),
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1, Toughness: 1}, Answers: game.AnswerPump},
			Effect: effects.Do(effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1,
				Label: "Cliffside Lookout — +1/+1"}),
		}},
	})
}
