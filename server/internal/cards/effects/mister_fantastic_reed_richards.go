package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mister Fantastic, Reed Richards — Legendary Creature — Human Scientist
// Hero {3}{U}, 2/4 (Marvel Super Heroes, #4455):
//
//	"Reach
//	 Whenever one or more tokens you control enter, you may draw a card."
//
// One trigger per batch of entries (Secure the Wastes' three tokens
// ask once; OncePerBatch, as Caretaker's Talent), then the CR 603.5
// "you may" prompt as it resolves. Mister Fantastic himself entering is
// not a token, and a token he sees enter alongside himself counts.
//
// No simplification.
func init() {
	const label = "Mister Fantastic — you may draw a card (tokens entered)"
	Register(Spec{
		OracleID:        "7fa5237e-dd79-4646-b935-cb8c6ee803ab",
		Name:            "Mister Fantastic, Reed Richards",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			OptionalDraw(OncePerBatch(On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsToken()
			}, label, Do(DrawCards{N: 1}))), "Mister Fantastic — draw a card?", 1),
		},
	})
}
