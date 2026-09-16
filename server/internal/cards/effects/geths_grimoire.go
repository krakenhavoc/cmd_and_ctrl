package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Geth's Grimoire — Artifact — Book {4} (EDHREC rank 2311):
//
//	"Whenever an opponent discards a card, you may draw a card."
//
// The discard deck's card-draw engine. Sangromancer's condition
// (b18OpponentDiscarded — the discarding player is the event's
// Actor) with "you may draw" as the payoff; the engine emits one
// discard event per card, so a wheel that makes an opponent discard
// seven asks seven times, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ef809e99-34a2-4471-8269-f56bf8037686",
		Name:         "Geth's Grimoire",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b18OpponentDiscarded(ev, source)
			}, "Geth's Grimoire — draw a card", Do(DrawCards{N: 1})), "Geth's Grimoire — draw a card?"),
		},
	})
}
