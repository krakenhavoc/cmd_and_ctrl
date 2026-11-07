package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rook Turret — Artifact Creature — Construct {3}{U}, 3/3 (Final
// Fantasy, #4462):
//
//	"Flying
//	 Whenever another artifact you control enters, you may draw a card.
//	 If you do, discard a card."
//
// Mechan Assembler's trigger condition (another artifact entering under
// your control, a token included) with Stadium Tidalmage's "you may draw
// a card. If you do, discard a card": one linked clause asked as the
// ability resolves. Declining the draw declines the discard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c5a6d752-9975-4059-8552-55c6ffb3ab68",
		Name:            "Rook Turret",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return artifactEnteredUnderYourControl(ev, source, g)
			}, "Rook Turret — you may draw a card; if you do, discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Rook Turret — draw a card, then discard a card?",
						Draws:    1,
						YesLabel: "Draw",
						NoLabel:  "Decline",
						Then:     celestusAnswer,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
