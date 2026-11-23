package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "78e65286-204d-43ab-8929-b8e2191b1fce",
		Name:            "Windrider Wizard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery(), effects.HasSubtype("Wizard")),
				"Windrider Wizard — you may draw a card; if you do, discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Windrider Wizard — draw a card, then discard a card?",
						Draws:    1,
						YesLabel: "Draw",
						NoLabel:  "Decline",
						Then:     effects.CelestusAnswer,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
