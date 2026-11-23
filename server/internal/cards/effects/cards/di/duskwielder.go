package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Duskwielder — Creature — Elf Berserker {B}, 1/2:
//
//	"Boast — {1}: Target opponent loses 1 life and you gain 1 life. (Activate only if
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
		OracleID:     "b9227f38-7db7-4ea1-9482-94e5de3984fe",
		Name:         "Duskwielder",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.BoastTargeting("{1}: Target opponent loses 1 life and you gain 1 life.",
				effects.ManaCost("{1}"), effects.TargetPlayer("target opponent", effects.Opponent()), effects.DrainTargetOpponentOne),
		},
	})
}
