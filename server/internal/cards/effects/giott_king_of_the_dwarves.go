package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Giott, King of the Dwarves — Legendary Creature — Dwarf Noble {R}{W},
// 1/1 (Final Fantasy, #4462):
//
//	"Double strike
//	 Whenever Giott or another Dwarf you control enters and whenever an
//	 Equipment you control enters, you may discard a card. If you do, draw
//	 a card."
//
// One ability with two conditions, so one trigger: a Dwarf you control
// entering (Giott is one, so "Giott or another Dwarf" is every Dwarf you
// control, the source included), or an Equipment you control entering.
// The Dwarf type is a word of Giott's text (CR 612.2) and the entering
// permanent's types are read as it is on the battlefield. Each
// permanent that enters triggers it separately. The rummage is Rubble
// Rouser's: the discard is optional and the draw only follows a discard.
//
// No simplification.
func init() {
	const label = "Giott, King of the Dwarves — you may discard a card; if you do, draw a card"
	Register(Spec{
		OracleID:        "c629fa91-e13c-494b-a95f-a91b7c52035a",
		Name:            "Giott, King of the Dwarves",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && (c.HasSubtype(source.TextSubtype("Dwarf")) || c.HasSubtype("Equipment"))
			}, label, func(g *game.Game, item *game.StackItem) error {
				return b39MayDiscardThenDraw(1, false, label, drawOneIfDiscarded)(NewContext(g, item))
			}),
		},
	})
}
