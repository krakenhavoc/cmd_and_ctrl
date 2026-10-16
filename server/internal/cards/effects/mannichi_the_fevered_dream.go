package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mannichi, the Fevered Dream — Legendary Creature — Spirit {2}{R}, 1/2:
//
//	"{1}{R}: Switch each creature's power and toughness until end of turn."
//
// The switch is layer 7d (CR 613.4d, #3996): it applies after every
// set, pump and counter, whatever their timestamps. Switching changes a characteristic,
// so the creatures are the ones on the battlefield as it resolves (CR
// 611.2c); a creature that enters later is not switched. It can kill
// a creature with 0 power, so the row answers as a pump and a removal.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3aafbf63-35c0-4f62-9281-a4d460e8d6f3",
		Name:         "Mannichi, the Fevered Dream",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}: Switch each creature's power and toughness until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerRemove},
			Cost:    ManaCost("{1}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SwitchPTUntilEOT{Match: Creature(), Label: "Mannichi, the Fevered Dream — switch each creature's power and toughness"}.Apply(NewContext(g, item))
			},
		}},
	})
}
