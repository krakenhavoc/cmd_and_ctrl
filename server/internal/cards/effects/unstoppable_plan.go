package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unstoppable Plan — Enchantment {2}{U} (EDHREC rank 1818):
//
//	"At the beginning of your end step, untap all nonland permanents
//	 you control."
//
// Pseudo-vigilance for the whole board, and a free second use of
// every tap ability. One "your end step" trigger (EventBeginEndStep
// gated on the active player being the controller) whose effect
// untaps every tapped nonland permanent the controller controls —
// b16UntapAllYouControlMatching, snapshot first so the untap events
// do not disturb the walk. Lands stay tapped, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3f6f4c98-ed7e-4fb1-8bfe-4210a39f77f2",
		Name:         "Unstoppable Plan",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginEndStep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Unstoppable Plan — untap all nonland permanents you control",
					func(g *game.Game, item *game.StackItem) error {
						return b16UntapAllYouControlMatching(NewContext(g, item), item.Controller, func(c game.Card) bool { return !c.IsLand() })
					})
			},
		}},
	})
}
