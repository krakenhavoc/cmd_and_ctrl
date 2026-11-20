package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "8fe1e619-f827-43f5-acaf-3b79aa671d1d",
		Name:         "Grisly Survivor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return discardedByYou(ev, source)
			}, "Grisly Survivor — +2/+0 until end of turn", thisGetsUntilEndOfTurn(2, 0, "Grisly Survivor — +2/+0 until end of turn")),
		},
	})
}
