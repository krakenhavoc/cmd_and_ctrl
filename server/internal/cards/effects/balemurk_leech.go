package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Balemurk Leech — Creature — Leech {1}{B}, 2/2:
//
//	"Eerie — Whenever an enchantment you control enters and whenever you
//	fully unlock a Room, each opponent loses 1 life."
//
// Life loss, not damage (eachOpponentLosesLife).
//
// Eerie is one ability with two conditions (Eerie, rooms.go): an
// enchantment entering under your control, or you fully unlocking a Room.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a511772c-3739-469d-a5a2-5827b7ce9b5c",
		Name:         "Balemurk Leech",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Eerie("Balemurk Leech — each opponent loses 1 life (eerie)", func(g *game.Game, item *game.StackItem) error {
				return eachOpponentLosesLife(g, item, 1)
			}),
		},
	})
}
