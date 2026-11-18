package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skyswimmer Koi — Creature — Fish {3}{U}, 3/3 (Kamigawa: Neon Dynasty,
// #5493):
//
//	"Flying
//	 Whenever an artifact you control enters, you may draw a card. If you
//	 do, discard a card."
//
// An enters trigger over any artifact under the Koi's controller (an
// artifact creature included; the Koi itself is not one): the optional
// loot is The Celestus' "you may draw, then discard" prompt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "53b9a2cf-2e6e-4ce0-bb9f-f33fb27b7b5e",
		Name:            "Skyswimmer Koi",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, neoB04ArtifactYouControlEntered,
				"Skyswimmer Koi — you may draw a card; if you do, discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Skyswimmer Koi — draw a card, then discard a card?",
						Draws:    1,
						YesLabel: "Draw",
						NoLabel:  "Decline",
						Then:     celestusAnswer,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
