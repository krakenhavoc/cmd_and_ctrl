package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Decorated Griffin — Creature — Griffin {4}{W}, 2/3 (Theros):
//
//	"Flying
//	 {1}{W}: Prevent the next 1 combat damage that would be dealt to you this
//	 turn."
//
// A one-point shield on the controller that only combat damage uses up
// (CR 615.7) and that ends with the turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8356b1fe-02f8-4bb9-bcdb-bb7454eb18ed",
		Name:            "Decorated Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}: Prevent the next 1 combat damage that would be dealt to you this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.ManaCost("{1}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.PreventNextDamage{Target: item.Controller, Amount: 1, CombatOnly: true, Label: "Decorated Griffin: prevent the next 1 combat damage"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
