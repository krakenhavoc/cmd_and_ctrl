package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Veteran of the Depths — Creature — Merfolk Soldier {3}{W}, 2/2 (Lorwyn, S82
// OLD, #5593):
//
//	"Whenever this creature becomes tapped, you may put a +1/+1 counter on it."
//
// Beckett Mariner's tap watch (ThisBecameTapped over the tap event and the attack
// that taps it, so a vigilance attacker does not trigger it), optional at
// resolution (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "70bcec56-1485-4e4e-8fff-940d7b133c5d",
		Name:         "Veteran of the Depths",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				"Veteran of the Depths — you may put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
				"Veteran of the Depths — put a +1/+1 counter on it?"),
		},
	})
}
