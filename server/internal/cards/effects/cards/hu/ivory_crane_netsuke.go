package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ivory Crane Netsuke — Artifact {2} (Saviors of Kamigawa, #5601):
//
//	"At the beginning of your upkeep, if you have seven or more cards in
//	 hand, you gain 4 life."
//
// An intervening if (CR 603.4): the hand is counted as the upkeep begins and
// again as the trigger resolves, so a hand that has shrunk below seven gains
// nothing.
//
// No simplification.
func init() {
	const label = "Ivory Crane Netsuke — you gain 4 life"
	effects.Register(effects.Spec{
		OracleID:     "ab151429-0890-4231-a3a7-34e8455dac70",
		Name:         "Ivory Crane Netsuke",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you have seven or more cards in hand", effects.OldB586YouHaveSevenCardsInHand,
				effects.TriggerWithPurpose(effects.AtYourUpkeep(label, effects.Do(effects.GainLife{Amount: 4})), game.Purpose{LifeGain: 4})),
		},
	})
}
