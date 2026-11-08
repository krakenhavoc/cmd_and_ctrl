package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Duelist of the Mind — Creature — Human Advisor {1}{U}, */3 (Outlaws of
// Thunder Junction, #4468):
//
//	"Flying, vigilance
//	 Duelist of the Mind's power is equal to the number of cards you've
//	 drawn this turn.
//	 Whenever you commit a crime, you may draw a card. If you do, discard a
//	 card. This ability triggers only once each turn."
//
// The power is a layer 7a characteristic-defining ability (CR 604.3,
// Collective Drone's shape) reading the controller's turn tally of cards
// drawn, which counts every draw this turn, the turn-based draw and the
// cards since discarded included. The crime trigger is Hardbristle
// Bandit's once-each-turn crime (CR 700.13) with Stadium Tidalmage's
// linked "you may draw a card. If you do, discard a card":
// declining the draw declines the discard. The loot's draw counts
// toward the power.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ff78f166-cca2-44bc-abe3-3a338bf543be",
		Name:            "Duelist of the Mind",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Static: []game.StaticAbility{{
			Label:     "Duelist of the Mind's power is equal to the number of cards you've drawn this turn.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = g.TurnTallyFor(source.Controller).CardsDrawn
			},
		}},
		Triggered: []game.TriggeredAbility{
			WheneverYouCommitACrimeOnceEachTurn("Duelist of the Mind — you may draw a card; if you do, discard a card (crime)",
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
