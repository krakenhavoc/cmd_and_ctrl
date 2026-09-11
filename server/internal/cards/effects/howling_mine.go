package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Howling Mine — Artifact {2}:
//
//	"At the beginning of each player's draw step, if this artifact is
//	 untapped, that player draws an additional card."
//
// A group-hug symmetrical draw engine, and the reason it is in a
// draw sprint rather than a politics one: the deck playing it has
// draw payoffs and the rest of the table does not.
//
// Two clauses worth being precise about.
//
// EACH PLAYER'S draw step, not yours — so the trigger fires four
// times a turn cycle at a four-player table, and the drawer is the
// event's Actor rather than the Mine's controller.
//
// IF THIS ARTIFACT IS UNTAPPED is a CR 603.4 intervening-if clause,
// checked twice: once when the trigger would go on the stack, and
// again on resolution. Tapping the Mine in response is the whole
// interaction (with Winter Orb, or a tapper) and it has to work, so
// the check is in AppliesTo *and* in the effect body.
//
// The trigger arrives after the turn-based draw has already happened
// — CR 504.1 does not use the stack — which is what "an ADDITIONAL
// card" means.
func init() {
	Register(Spec{
		OracleID: "d26b27db-a567-4631-b4b6-7294222fbdd1",
		Name:     "Howling Mine",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginDrawStep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				// Intervening-if, first check (CR 603.4): a tapped
				// Mine never puts the trigger on the stack at all.
				return !source.Tapped
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				// The DRAWER is the player whose draw step it is, not
				// the Mine's controller. Captured by value, which is
				// clone-safe.
				drawer := ev.Actor
				return game.NewTriggeredItem(source, "Howling Mine — draw an additional card",
					func(g *game.Game, item *game.StackItem) error {
						// Intervening-if, second check (CR 603.4):
						// tapping the Mine in response is the play,
						// and it has to stop the draw.
						if c, ok := g.LookupCardForEffect(item.SourceCardID); !ok || c.Tapped {
							return nil
						}
						return g.DrawNForEffect(drawer, 1)
					})
			},
		}},
	})
}
