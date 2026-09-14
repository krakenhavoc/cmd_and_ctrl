package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tegwyll, Duke of Splendor — Legendary Creature — Faerie Noble
// {1}{U}{B}, 2/3 (EDHREC rank 3827):
//
//	"Flying, deathtouch
//	 Other Faeries you control get +1/+1.
//	 Whenever another Faerie you control dies, you draw a card and you
//	 lose 1 life."
//
// The Faerie lord with a Grim Haruspex stapled on. Both keywords
// ride PrintedKeywords; the anthem is the shared TribalAnthem
// builder with "other" and "you control"; the dies trigger fires for
// every OTHER Faerie the controller controlled that goes to a
// graveyard — Tegwyll's own death does not count, as printed — and
// the draw comes before the life loss, in printed order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "868f0a0a-ca9e-4baf-8295-6b228aa834e5",
		Name:            "Tegwyll, Duke of Splendor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Faerie"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b36AnotherFaerieYouControlDied(ev, source, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Tegwyll, Duke of Splendor — draw a card and lose 1 life", b36DrawAndLoseOne)
			},
		}},
	})
}
