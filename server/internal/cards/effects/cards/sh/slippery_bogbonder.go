package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slippery Bogbonder — Creature — Human Druid {3}{G}, 3/3:
//
//	"Flash
//	 Hexproof
//	 When this creature enters, put a hexproof counter on target creature.
//	 Then move any number of counters from among creatures you control onto
//	 that creature."
//
// The target is any creature (an opponent's too), chosen as the trigger goes
// on the stack and re-checked as it resolves; one that is gone fizzles the
// whole ability (CR 608.2b). The hexproof counter goes on first, then each
// other creature of the controller's that carries counters is asked in turn
// how many of each kind to move (CR 608.2c, the Resourceful Defense
// question), and they move through the same counter primitive as a
// placement, so a doubler sees each one. The target itself is never a
// source of the moves, and a creature that left in the meantime is skipped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2f946c13-cfab-415f-9df7-6d621761b2fc",
		Name:            "Slippery Bogbonder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "hexproof"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Slippery Bogbonder — put a hexproof counter on target creature, then move any number of counters from among creatures you control onto it",
					effects.RprB23BogbonderResolve),
				effects.TargetCreature("target creature")),
		},
	})
}
