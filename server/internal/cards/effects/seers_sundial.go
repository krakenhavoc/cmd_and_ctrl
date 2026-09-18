package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seer's Sundial — Artifact for {4} (EDHREC rank 4413):
//
//	"Landfall — Whenever a land you control enters, you may pay
//	 {2}. If you do, draw a card."
//
// Colourless, repeatable card draw priced in mana you often have
// spare after a land drop. In a landfall deck with fetchlands and
// extra land drops it is a Howling Mine that only you get to use.
// Roadmap batch 42 (#449), "no new machinery".
//
// "You may pay {2}. If you do, draw" is a payment, not a prompt with
// a free out: MayPay puts the offer to the controller when the
// trigger resolves, and the draw happens only when the mana is
// actually paid. Declining costs nothing.
//
// LANDFALL fires on a land you control ENTERING, however it got
// there — a land drop, a Cultivate, a Crucible of Worlds replay, an
// opponent's Explosive Vegetation that put a land under your control.
// It does not fire on a land you play that is countered or on a land
// entering under somebody else's control.
//
// One trigger per land. Two lands entering at once from one
// Scapeshift give two triggers and two separate {2} offers, which is
// the printed card — "whenever A land" is per-land, unlike the "one
// or more" batch wording.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9929d7ba-cdc5-4099-a0ee-3a15073336f3",
		Name:         "Seer's Sundial",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("Seer's Sundial — pay {2} to draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}",
						Question: "Seer's Sundial — pay {2} to draw a card?",
						OnPay: func(ctx *Context) error {
							return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
						},
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
