package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lavafume Invoker — Creature — Goblin Shaman {2}{R}, 2/2 (Rise of the
// Eldrazi, S82 OLD batch 483, #5581):
//
//	"{8}: Creatures you control get +3/+0 until end of turn."
//
// Behind the Scenes' team pump: the set is the creatures its controller
// controls as the ability resolves (CR 611.2c), and it is a layer 7c effect.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "41b49a8e-09c5-4fd1-95f9-e5a51ce24f69",
		Name:         "Lavafume Invoker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{8}: Creatures you control get +3/+0 until end of turn.",
			Cost:    effects.ManaCost("{8}"),
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 3}, Answers: game.AnswerPump},
			Effect: effects.Do(effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.YouControl()), Power: 3,
				Label: "Lavafume Invoker — creatures you control get +3/+0",
			}),
		}},
	})
}
