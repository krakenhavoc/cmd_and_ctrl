package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Revered Elder — Creature — Human Cleric {2}{W}, 1/2 (Mercadian Masques, S82
// OLD b682, #5619):
//
//	"{1}: Prevent the next 1 damage that would be dealt to this creature this
//	 turn."
//
// Ordruun Commando's shield for {1} (PreventNextDamage, CR 615.7), on the
// Elder itself while it is still the same object (CR 400.7), as often as it
// can be paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e8d18dc7-f44b-47f6-8400-3f563c2f19b7",
		Name:         "Revered Elder",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: Prevent the next 1 damage that would be dealt to this creature this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.ManaCost("{1}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.PreventNextDamage{Target: item.SourceCardID, Amount: 1,
					Label: "Revered Elder: prevent the next 1 damage"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
