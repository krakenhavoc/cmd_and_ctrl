package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mundungu — Creature — Human Wizard {1}{U}{B}, 1/1 (Visions, S82 older sets
// batch 791, #5631):
//
//	"{T}: Counter target spell unless its controller pays {1} and 1 life."
//
// A soft counter whose price is one payment of mana and life together (CR
// 118.12): the spell's controller is asked as the ability resolves, a payer who
// can't fund both declines, and a spell that has left the stack by then is
// untouched (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3fb4161d-a5b9-411d-942c-2fdb039e7bde",
		Name:         "Mundungu",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Counter target spell unless its controller pays {1} and 1 life.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetSpell("target spell"),
			Effect:  effects.Old791CounterUnlessPaysManaAndLife("Mundungu — pay {1} and 1 life, or your spell is countered"),
		}},
	})
}
