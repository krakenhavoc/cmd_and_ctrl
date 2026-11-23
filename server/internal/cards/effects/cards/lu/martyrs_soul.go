package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Martyr's Soul — Creature — Spirit Soldier {2}{W}, 3/2 (Jumpstart 2022):
//
//	"Convoke
//	 When this creature enters, if you control no tapped lands, put two
//	 +1/+1 counters on it."
//
// Convoke is Spec.TapCost (Flockchaser Phantom). The trigger is an
// intervening "if" (CR 603.4), checked as it fires and again as it resolves
// (CondNoTappedLands), so a land tapped in response leaves it doing nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "886382dc-5284-4b84-ad8c-8d8463d42fa0",
		Name:         "Martyr's Soul",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondNoTappedLands{},
				effects.WhenThisEnters("Martyr's Soul — put two +1/+1 counters on it",
					effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 2}))),
		},
	})
}
