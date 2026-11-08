package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Projektor Inspector — Creature — Human Detective {2}{U}, 3/2 (Murders
// at Karlov Manor, #4469):
//
//	"Whenever this creature or another Detective you control enters and
//	 whenever a Detective you control is turned face up, you may draw a
//	 card. If you do, discard a card."
//
// One ability on two event kinds (ADetectiveEnteredOrWasTurnedFaceUp, its
// own entry included). The "may" is Stadium Tidalmage's loot: a yes/no as
// the ability resolves, then draw one and discard one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "58b27a79-ba63-45d4-84ee-423ce0d91f6b",
		Name:         "Projektor Inspector",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventTurnedFaceUp}, ADetectiveEnteredOrWasTurnedFaceUp(true),
				"Projektor Inspector — you may draw a card; if you do, discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Projektor Inspector — draw a card, then discard a card?",
						Draws:    1,
						YesLabel: "Draw",
						NoLabel:  "Decline",
						Then:     celestusAnswer,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
