package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Count on Luck — Enchantment {R}{R}{R} (EDHREC rank 4276):
//
//	"At the beginning of your upkeep, exile the top card of your
//	 library. You may play that card this turn."
//
// Three red pips for a free card a turn, in the only colour that has
// to pay for its card advantage by using it immediately. The tension
// is the whole card: the exiled card is gone at end of turn whether or
// not you could afford it, so a Count on Luck deck wants a low curve
// and a lot of lands.
//
// "PLAY that card", not "cast" it — a land exiled this way can be
// played as that turn's land drop, which is the difference between
// this and an impulse-draw that says "cast". The permission is the
// same ExilePlayPermission every impulse card rides, so the client's
// existing exile-zone button renders it with no new code.
//
// It is the CONTROLLER's upkeep, so once a turn cycle rather than
// once per player's turn, and the exile happens from the controller's
// own library.
//
// The card is exiled face UP. Everyone at the table sees what you got,
// which is the printed behaviour and occasionally matters — an
// opponent holding a counterspell knows whether to keep it up.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "73ad7bcd-4ffc-442c-b6b5-d935b9ccaaaf",
		Name:         "Count on Luck",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Count on Luck — exile the top card; you may play it this turn",
				func(g *game.Game, item *game.StackItem) error {
					_, err := b12ImpulseExileForTurn(g, item, 1)
					return err
				}),
		},
	})
}
