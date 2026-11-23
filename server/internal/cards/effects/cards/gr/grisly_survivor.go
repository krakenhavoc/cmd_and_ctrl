package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grisly Survivor — Creature — Minotaur Warrior {2}{B}, 2/3 (Hour of
// Devastation, #5534):
//
//	"Whenever you cycle or discard a card, this creature gets +2/+0 until
//	 end of turn."
//
// "Cycle or discard" is one event, the discard: cycling's cost is to
// discard the card (CR 702.29d), so a cycled card is one trigger, not two
// (Archfiend of Ifnir). The pump is the shared thisGetsUntilEndOfTurn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8fe1e619-f827-43f5-acaf-3b79aa671d1d",
		Name:         "Grisly Survivor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.DiscardedByYou(ev, source)
			}, "Grisly Survivor — +2/+0 until end of turn", effects.ThisGetsUntilEndOfTurn(2, 0, "Grisly Survivor — +2/+0 until end of turn")),
		},
	})
}
