package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gerrard's Battle Cry — Enchantment {W} (Tempest, S82 OLD b755, #5628):
//
//	"{2}{W}: Creatures you control get +1/+1 until end of turn."
//
// Behind the Scenes' team pump: the creatures are the ones you control as the
// ability resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "12ef651c-5a86-49c4-a512-292e9d135a76",
		Name:         "Gerrard's Battle Cry",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}: Creatures you control get +1/+1 until end of turn.",
			Cost:    effects.ManaCost("{2}{W}"),
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1, Toughness: 1}, Answers: game.AnswerPump},
			Effect: effects.Do(effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1,
				Label: "Gerrard's Battle Cry — creatures you control get +1/+1",
			}),
		}},
	})
}
