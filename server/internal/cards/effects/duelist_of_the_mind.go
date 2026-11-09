package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Duelist of the Mind — Creature — Human Advisor {1}{U}, */3 (Outlaws of Thunder Junction):
//
//	"Flying, vigilance
//	 Duelist of the Mind's power is equal to the number of cards you've drawn
//	 this turn.
//	 Whenever you commit a crime, you may draw a card. If you do, discard a
//	 card. This ability triggers only once each turn."
//
// The power is a layer 7a characteristic-defining ability over the turn's
// draw tally (every draw counts, a card since discarded included), which
// is why the static declares DependsOnEvents{EventDrawCard}: a draw moves
// nothing on the battlefield, so without the hint the power would stay stale
// until something unrelated invalidated it. The crime trigger is held to
// once each turn, and its loot is the Stadium Tidalmage clause (a "may draw"
// whose discard follows only if the draw happened).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ff78f166-cca2-44bc-abe3-3a338bf543be",
		Name:            "Duelist of the Mind",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Static: []game.StaticAbility{{
			Label:           "Duelist of the Mind's power is equal to the number of cards you've drawn this turn.",
			Layer:           game.Layer7PT,
			SubLayer:        game.SubLayer7A_CDA,
			AppliesTo:       selfOnly,
			DependsOnEvents: []game.EventKind{game.EventDrawCard},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = g.TurnTallyFor(source.Controller).CardsDrawn
			},
		}},
		Triggered: []game.TriggeredAbility{
			WheneverYouCommitACrimeOnceEachTurn("Duelist of the Mind — you may draw a card; if you do, discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Duelist of the Mind — draw a card, then discard a card?",
						Draws:    1,
						YesLabel: "Draw",
						NoLabel:  "Decline",
						Then:     celestusAnswer,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
