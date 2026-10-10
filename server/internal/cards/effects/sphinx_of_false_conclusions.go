package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sphinx of False Conclusions — Creature — Sphinx Illusion {2}{U}{U},
// 4/2:
//
//	"Flash
//	 Flying
//	 Whenever this creature attacks, draw a card, then discard a card.
//	 When this creature dies, if it isn't a token, create a token that's
//	 a copy of it."
//
// The dies half is Vaultborn Tyrant's shape without the "except": the
// token check reads the dying object's last-known supertypes (CR
// 603.10), so the copy is a token and its own death makes nothing. The
// loot draws first and opens the discard once every draw has landed, so
// the drawn card is a legal discard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9f004382-c57b-4fde-af97-0c85516c3cf4",
		Name:            "Sphinx of False Conclusions",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Sphinx of False Conclusions — draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error { return lootOne(g, item, 1) }),
			{
				Watches: []game.EventKind{game.EventLTB},
				Key:     "Sphinx of False Conclusions — create a token copy of it",
				AppliesTo: func(ev game.Event, source *game.Card, lki game.Characteristic, _ *game.Game) bool {
					return cardDied(ev, source) && !hasFold(lki.Supertypes, "Token")
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateTokenCopy{Controller: item.Controller, Copy: item.SourceCardID, N: 1}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
