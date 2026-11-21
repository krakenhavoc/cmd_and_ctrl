package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fearless Liberator — Creature — Dwarf Berserker {1}{R}, 2/1:
//
//	"Boast — {2}{R}: Create a 2/1 red Dwarf Berserker creature token. (Activate only if
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
		OracleID:     "b0b93253-6432-401e-9a01-94c41fb72c10",
		Name:         "Fearless Liberator",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.BoastAnswering(game.AnswerMakesBlocker, "{2}{R}: Create a 2/1 red Dwarf Berserker creature token.",
				effects.ManaCost("{2}{R}"), effects.CreateTheToken("2/1 red Dwarf Berserker")),
		},
	})
}
