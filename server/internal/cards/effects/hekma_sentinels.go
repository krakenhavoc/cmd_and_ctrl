package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hekma Sentinels — Creature — Human Cleric {2}{U}, 2/3
// (Amonkhet, #5535):
//
//	"Whenever you cycle or discard a card, this creature gets +1/+1 until end of turn."
//
// "Cycle or discard" is one event, the discard: cycling's cost discards
// the card (CR 702.29d), so watching the cycle as well would count one
// cycle twice (see Curator of Mysteries).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "90ea758d-3783-47a5-bb18-9ff9377113fb",
		Name:         "Hekma Sentinels",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return discardedByYou(ev, source)
			}, "Hekma Sentinels — it gets +1/+1 until end of turn", thisGetsUntilEndOfTurn(1, 1, "Hekma Sentinels — +1/+1")),
		},
	})
}
