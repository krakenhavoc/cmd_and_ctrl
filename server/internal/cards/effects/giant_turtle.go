package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Giant Turtle — Creature — Turtle {1}{G}{G}, 2/4:
//
//	"This creature can't attack if it attacked during your last turn."
//
// A "can't attack" restriction (CR 508.1c) that applies while ADR 0108 §6's
// record says this creature was declared as an attacker during its
// controller's last turn. Per the Giant Turtle ruling it only cares about
// YOUR last turn, not an opponent's: in a four-player game the record is
// kept through three other players' turns. A creature put onto the
// battlefield attacking never "attacked" (CR 508.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9297c0a6-1a8e-4e6e-99d6-f0877b2ec46c",
		Name:         "Giant Turtle",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{CantAttackIfItAttackedDuringYourLastTurn()},
	})
}
