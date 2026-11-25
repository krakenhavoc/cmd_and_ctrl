package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ensnaring Bridge — Artifact {3}:
//
//	"Creatures with power greater than the number of cards in your hand
//	 can't attack."
//
// A CR 508.1c restriction on every creature on the battlefield, its
// controller's own included (#6087, game.AttackTargetRestriction.
// PowerAtMostHandOf). "Your hand" is the Bridge controller's, and both the
// attacker's power (layers and counters) and that hand's size are read live
// when attackers are declared, so a card drawn or discarded, or a pump cast,
// before the declaration counts. Power equal to the hand size may attack.
// Once declared, a creature keeps attacking whatever happens to the hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dade5835-ccaf-45e3-b9a3-46017e0384b3",
		Name:         "Ensnaring Bridge",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.CreaturesWithPowerAboveYourHandCantAttack("Creatures with power greater than the number of cards in your hand can't attack."),
		},
	})
}
