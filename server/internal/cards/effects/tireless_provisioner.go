package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tireless Provisioner — Creature — Elf Scout {2}{G}, 3/2 (EDHREC
// rank 184):
//
//	"Landfall — Whenever a land you control enters, create a Food
//	 token or a Treasure token."
//
// A ramp creature that turns every land drop into an extra mana —
// or, when the mana isn't needed, three life. Landfall is an ETB
// trigger filtered to lands, and it fires for any way a land enters
// under your control: played, fetched, put onto the battlefield by
// a spell. The fetchland case fires twice, once for the fetch land
// and once for the basic it finds, as in paper.
//
// Sandbox simplification: ALWAYS a Treasure. The choice between the
// two tokens is made when the trigger resolves, and the engine has
// no resolution-time option prompt for a trigger (PendingChoice has
// no "pick one of N labelled options" kind; the modal machinery is
// cast-time only). Treasure is the pick a player makes in nearly
// every game and the one the card is played for, and since the
// printed card always allows it, this is never stronger than
// printed — only less flexible. The day a resolution-time option
// prompt lands, this trigger should offer FoodToken() as the other
// answer.
func init() {
	Register(Spec{
		OracleID: "ab8d5f5c-1976-4f77-8ed2-8d28ee666741",
		Name:     "Tireless Provisioner",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsLand()
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Tireless Provisioner — create a Treasure (landfall)",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{
							Controller: item.Controller,
							Template:   TreasureToken(),
							N:          1,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
