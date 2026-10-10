package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trained Pronghorn — Creature — Antelope {1}{W}, 1/1:
//
//	"Discard a card: Prevent all damage that would be dealt to this
//	 creature this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the not-one-use shield pinned to
// this creature. A Pronghorn that has left and come back before the
// ability resolves is a new object and is not shielded (CR 400.7).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "22528380-4f47-4bd9-a1d4-79f700cba4f5",
		Name:         "Trained Pronghorn",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Discard a card: Prevent all damage that would be dealt to this creature this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    DiscardACard(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return shieldThis(g, item, PreventDamageFromSource{})
			},
		}},
	})
}
