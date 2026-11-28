package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ethereal Champion — Creature — Avatar {2}{W}{W}{W}, 3/4 (Mirage, S82 OLD
// batch 809, #5633):
//
//	"Pay 1 life: Prevent the next 1 damage that would be dealt to this
//	 creature this turn."
//
// Ordruun Commando's shield (CR 615.7) with a life cost instead of mana.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f203671b-76f7-4be0-80a9-826fbb450fe7",
		Name:         "Ethereal Champion",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Pay 1 life: Prevent the next 1 damage that would be dealt to this creature this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.PayLife(1),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.PreventNextDamage{Target: item.SourceCardID, Amount: 1,
					Label: "Ethereal Champion: prevent the next 1 damage"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
