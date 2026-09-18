package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Patron of the Arts — Creature — Dragon Noble {2}{R}, 3/1 (EDHREC
// rank 4198):
//
//	"When this creature enters or dies, create a Treasure token.
//	 (It's an artifact with "{T}, Sacrifice this token: Add one mana
//	 of any color.")"
//
// A three-mana 3/1 that refunds a mana on the way in and another on
// the way out, which is the shape a Treasure deck wants: it is a
// Dragon for Miirym and Sarkhan, a body for Goldspan Dragon's tribe,
// and two artifacts-entering triggers for an Academy Manufactor or a
// Reckless Fireweaver.
//
// "Enters or dies" is ONE printed ability with two trigger
// conditions, so it is one TriggeredAbility watching two event kinds
// — Stitcher's Supplier's shape. EventLTB is gated by cardDied, so a
// bounce, a flicker's exile leg or a tuck to the library does not pay
// out; only an actual battlefield-to-graveyard move does.
//
// The Treasure is the catalog's behavioural token (TreasureToken
// carries its own sacrifice-for-mana ability), not a table row, so
// the token this makes is a real mana source rather than a blank
// artifact.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "47f8df80-abc9-4d46-8078-8fbc23430259",
		Name:         "Patron of the Arts",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventLTB}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				if ev.Kind == game.EventETB {
					return ev.CardID == source.InstanceID
				}
				return cardDied(ev, source)
			}, "Patron of the Arts — create a Treasure token", Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}
