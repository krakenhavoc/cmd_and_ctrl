package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Genesis Chamber — Artifact {2} (EDHREC rank 2818):
//
//	"Whenever a nontoken creature enters, if this artifact is
//	 untapped, that creature's controller creates a 1/1 colorless Myr
//	 artifact creature token."
//
// The symmetric token engine. Any player's nontoken creature — a
// cast, a reanimation, a flicker's return — hands ITS CONTROLLER a
// Myr; the Myr is a token, so it never re-triggers. "If this artifact
// is untapped" is an intervening-if (CR 603.4): checked when the
// creature enters, and again as the trigger resolves, so tapping the
// Chamber in response (a Voltaic Key, a Tumble Magnet) stops the
// Myr. The controller is read off the entering creature at trigger
// time and carried on the item as a value.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "150ab025-5cc3-4468-a724-bbba8838445d",
		Name:         "Genesis Chamber",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if source.Tapped {
					return false
				}
				_, ok := b26NontokenCreatureEntered(ev, g)
				return ok
			},
			Key: "Genesis Chamber — that creature's controller creates a Myr",
			// The entering creature's controller is a board read at
			// trigger time (ADR 0041 P9's fill-in Build), carried as
			// item.Params.Player.
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				entered, _ := b26NontokenCreatureEntered(ev, g)
				item := game.NewTriggeredItem(source, "Genesis Chamber — that creature's controller creates a Myr")
				item.Params.Player = entered.Controller
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				controller := item.Params.Player
				if !b26SourceOnBattlefieldUntapped(g, item) {
					return nil
				}
				if g.PlayerByIDForEffect(controller) == nil {
					return nil
				}
				return CreateToken{Controller: controller, Template: TokenCard("1/1 colorless Myr artifact"), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
