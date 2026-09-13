package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Font of Mythos — Artifact {4} (EDHREC rank 2159):
//
//	"At the beginning of each player's draw step, that player draws
//	 two additional cards."
//
// Howling Mine's bigger sibling, without the "if untapped" clause.
// EACH PLAYER'S draw step, so the trigger fires once per seat per
// round and the drawer is the event's Actor — the player whose draw
// step it is — never the Font's controller. The trigger arrives
// after the turn-based draw (CR 504.1 does not use the stack), which
// is what "additional" means.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7194a262-5e7a-4c12-b271-2bc5e0799477",
		Name:         "Font of Mythos",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginDrawStep},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				drawer := ev.Actor
				return game.NewTriggeredItem(source, "Font of Mythos — draw two additional cards",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: drawer, N: 2}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
