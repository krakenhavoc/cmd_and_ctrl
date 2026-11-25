package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rafter Demon — Creature — Demon {2}{B}{R}, 4/2 (Ravnica Allegiance):
//
//	"Spectacle {3}{B}{R} (You may cast this spell for its spectacle cost
//	 rather than its mana cost if an opponent lost life this turn.)
//	 When this creature enters, if its spectacle cost was paid, each opponent
//	 discards a card."
//
// Spectacle is the shared alternative cost; "if its spectacle cost was paid"
// is an intervening if over the permanent's cast provenance (Latchkey
// Faerie's shape), so a hard-cast Demon makes nobody discard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "2e7a03ac-44bb-45f1-b209-85fc3b6e90d3",
		Name:             "Rafter Demon",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Spectacle("{3}{B}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, castWithSpectacle),
				"Rafter Demon — each opponent discards a card",
				func(g *game.Game, item *game.StackItem) error {
					effects.EachOpponentDiscards(effects.NewContext(g, item), 1)
					return nil
				}),
		},
	})
}

// castWithSpectacle is "if its spectacle cost was paid" on a permanent: its cast
// provenance names the spectacle cost (CR 400.7d, CR 702.137a).
func castWithSpectacle(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source != nil && source.Provenance.AltCost == "spectacle"
}
