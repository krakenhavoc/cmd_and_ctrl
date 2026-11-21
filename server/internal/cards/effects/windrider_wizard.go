package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Windrider Wizard — Creature — Human Wizard {2}{U}, 2/2 (Commander Legends:
// Battle for Baldur's Gate):
//
//	"Flying
//	 Whenever you cast an instant, sorcery, or Wizard spell, you may draw a
//	 card. If you do, discard a card."
//
// Stadium Tidalmage's optional draw-then-discard (MayChoice with the draw, the
// discard as the "if you do" tail), on a cast of an instant, a sorcery or a
// Wizard spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "78e65286-204d-43ab-8929-b8e2191b1fce",
		Name:            "Windrider Wizard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery(), HasSubtype("Wizard")),
				"Windrider Wizard — you may draw a card; if you do, discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Windrider Wizard — draw a card, then discard a card?",
						Draws:    1,
						YesLabel: "Draw",
						NoLabel:  "Decline",
						Then:     celestusAnswer,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
