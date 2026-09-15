package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stitcher's Supplier — Creature — Zombie {B}, 1/1 (EDHREC rank 622):
//
//	"When this creature enters or dies, mill three cards."
//
// The graveyard deck's one-drop: three cards in the yard on the way
// in and three more on the way out. "Enters or dies" is ONE printed
// ability with two trigger conditions, so it is one TriggeredAbility
// watching two event kinds (Sun Titan's "enters or attacks" is the
// precedent) — EventETB for the entry, and EventLTB gated by cardDied
// so a bounce or an exile does not count as dying.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7fd61a18-6e4f-40c5-aa00-3d101ec1ec82",
		Name:         "Stitcher's Supplier",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventLTB}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				if ev.Kind == game.EventETB {
					return ev.CardID == source.InstanceID
				}
				return cardDied(ev, source)
			}, "Stitcher's Supplier — mill three cards", Do(MillCards{N: 3})),
		},
	})
}
