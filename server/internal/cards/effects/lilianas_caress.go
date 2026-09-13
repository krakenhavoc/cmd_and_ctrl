package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Liliana's Caress — Enchantment {1}{B} (EDHREC rank 2412):
//
//	"Whenever an opponent discards a card, that player loses 2 life."
//
// The discard deck's Megrim. Sangromancer's condition — the
// discarding player is the event's Actor — with the loss going to
// that player, captured in Build. Once per card discarded, so a
// wheel that makes an opponent discard seven is fourteen life, as
// printed. Life loss, not damage: no prevention or doubler sees it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a4aec0d6-13fa-4709-b1a9-2f483f032744",
		Name:         "Liliana's Caress",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDiscardCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b18OpponentDiscarded(ev, source)
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				victim := ev.Actor
				return game.NewTriggeredItem(source, "Liliana's Caress — that player loses 2 life",
					func(g *game.Game, item *game.StackItem) error {
						return g.ChangePlayerLifeForEffect(item.SourceCardID, victim, -2)
					})
			},
		}},
	})
}
