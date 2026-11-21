package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Usher of the Fallen — Creature — Spirit Warrior {W}, 2/1:
//
//	"Boast — {1}{W}: Create a 1/1 white Human Warrior creature token. (Activate only if
//	 this creature attacked this turn and only once each turn.)"
//
// Boast (CR 702.142a) is built with the Boast constructor (boast.go):
// the engine reads the attack record and the activation tally, so the
// card names neither. The activation is spent at the announce, whether
// or not it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6eb06dc-a62d-4fdb-b336-e304d8d68c92",
		Name:         "Usher of the Fallen",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.BoastAnswering(game.AnswerMakesBlocker, "{1}{W}: Create a 1/1 white Human Warrior creature token.",
				effects.ManaCost("{1}{W}"), effects.CreateTheToken("1/1 white Human Warrior")),
		},
	})
}
