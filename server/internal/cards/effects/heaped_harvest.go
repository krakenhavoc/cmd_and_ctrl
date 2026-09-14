package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heaped Harvest — Artifact — Food {2}{G} (EDHREC rank 2796):
//
//	"When this artifact enters and when you sacrifice it, you may
//	 search your library for a basic land card, put it onto the
//	 battlefield tapped, then shuffle.
//	 {2}, {T}, Sacrifice this artifact: You gain 3 life."
//
// A Food that ramps twice. The search is one printed ability with
// two conditions — the entry and the controller's own sacrifice of
// it — so one declaration watches both kinds; EventSacrifice fires
// while the Harvest is still on the battlefield, which is how the
// harvester finds it. "You may" is the search prompt's decline, and
// the land enters tapped through the search's own clause. The Food
// ability is the ordinary one, with the sacrifice as its cost — so
// cracking it queues the search above the life gain, and the life
// lands after the land does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bbfc5011-a9b7-442d-a443-974a5a64de46",
		Name:         "Heaped Harvest",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB, game.EventSacrifice},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b26SelfEnteredOrWasSacrificed(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Heaped Harvest — you may search for a basic land, tapped",
					func(g *game.Game, item *game.StackItem) error {
						return b07SearchBasicOntoBattlefield(g, item, item.Controller, true, true, "Heaped Harvest — a basic land card, onto the battlefield tapped")
					})
			},
		}},
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}, Sacrifice this artifact: You gain 3 life.",
			Cost:  Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GainLife{Player: item.Controller, Amount: 3}.Apply(NewContext(g, item))
			},
		}},
	})
}
