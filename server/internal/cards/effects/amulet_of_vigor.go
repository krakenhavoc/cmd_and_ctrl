package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Amulet of Vigor — Artifact {1} (EDHREC rank 1280):
//
//	"Whenever a permanent you control enters tapped, untap it."
//
// The bounce-land and Overlook enabler. "Enters tapped" is read off
// the permanent at the moment EventETB fires: every entry path in the
// engine — the spell resolving, the land play, the library search,
// a return from exile, a token made tapped — stamps Tapped before it
// announces the entry, so a permanent that is tapped when the event
// arrives entered that way. The untap is a trigger with a response
// window, as printed, and does nothing if the permanent has left or
// been untapped by then. The Amulet sees its own tapped entry too,
// which is the printed reading.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dba16032-66c1-4ccb-9d65-d41ac550d182",
		Name:         "Amulet of Vigor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.Tapped
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				entered := ev.CardID
				return game.NewTriggeredItem(source, "Amulet of Vigor — untap it",
					func(g *game.Game, item *game.StackItem) error {
						if !onBattlefield(g, entered) {
							return nil
						}
						if c, ok := g.LookupCardForEffect(entered); !ok || !c.Tapped {
							return nil
						}
						return UntapTarget{Target: entered}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
