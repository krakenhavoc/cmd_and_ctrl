package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Akroan Phalanx — Creature — Human Soldier {3}{W}, 3/3 (Born of the Gods, S82
// OLD b378, #5559):
//
//	"Vigilance
//	 {2}{R}: Creatures you control get +1/+0 until end of turn."
//
// Ruby Collector's team pump: the creatures you control as it resolves
// (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0edd4894-0df0-4a4e-8a96-d20360ea81c7",
		Name:            "Akroan Phalanx",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}: Creatures you control get +1/+0 until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1}, Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1,
					Label: "Akroan Phalanx — creatures you control get +1/+0",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
