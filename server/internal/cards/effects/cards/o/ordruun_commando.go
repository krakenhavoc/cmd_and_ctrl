package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ordruun Commando — Creature — Minotaur Soldier {3}{R}, 4/1 (Ravnica: City of
// Guilds, S82 OLD batch 574, #5600):
//
//	"{W}: Prevent the next 1 damage that would be dealt to this creature this
//	 turn."
//
// A one-point prevention shield on the Commando for the turn (CR 615.7), the
// charged shield Kitsune Healer uses, written only while the Commando is
// still the permanent that was activated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f44a8a72-2c51-429e-8447-7bd8155b6f72",
		Name:         "Ordruun Commando",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}: Prevent the next 1 damage that would be dealt to this creature this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.ManaCost("{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.PreventNextDamage{Target: item.SourceCardID, Amount: 1,
					Label: "Ordruun Commando: prevent the next 1 damage"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
