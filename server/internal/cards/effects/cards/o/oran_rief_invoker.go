package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oran-Rief Invoker — Creature — Human Shaman {1}{G}, 2/2 (Battle for
// Zendikar):
//
//	"{8}: This creature gets +5/+5 and gains trample until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c5dcab94-f4c0-4e5e-ad99-7b3871be6e89",
		Name:         "Oran-Rief Invoker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{8}: This creature gets +5/+5 and gains trample until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 5, Toughness: 5}},
			Cost:    effects.ManaCost("{8}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := effects.ThisGetsUntilEndOfTurn(5, 5, "Oran-Rief Invoker — +5/+5 until end of turn")(g, item); err != nil {
					return err
				}
				return effects.ThisGainsKeywordUntilEndOfTurn("trample", "Oran-Rief Invoker — trample until end of turn")(g, item)
			},
		}},
	})
}
