package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cunning Survivor — {1}{U} Creature — Human Warrior 1/3 (Hour of Devastation):
//
//	"Whenever you cycle or discard a card, this creature gets +1/+0 until end
//	 of turn and can't be blocked this turn."
//
// "Cycle or discard" is one event, the discard: cycling's cost is discarding
// the card (CR 702.29d), so watching the discard alone catches a cycled card and
// every other discard once each (Archfiend of Ifnir's note).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6dfac7c-72ed-4db8-97b0-371ed9108b00",
		Name:         "Cunning Survivor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.DiscardedByYou(ev, source)
			}, "Cunning Survivor — it gets +1/+0 until end of turn and can't be blocked this turn",
				effects.ThisGetsPlusOneZeroAndCantBeBlockedThisTurn("Cunning Survivor")),
		},
	})
}
