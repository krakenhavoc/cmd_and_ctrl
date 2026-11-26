package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Screeching Phoenix — Creature — Phoenix {4}{R}{R}, 4/4 (Global Series Jiang Yanggu & Mu Yanling, #5528):
//
//	"Flying"
//	"{2}{R}: Creatures you control get +1/+0 until end of turn."
//
// Purphoros's team pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2b75759a-c3cc-432d-9609-4f8fa2456870",
		Name:            "Screeching Phoenix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}: Creatures you control get +1/+0 until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1}, Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl()),
					Power: 1,
					Label: "Screeching Phoenix — creatures you control get +1/+0",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
