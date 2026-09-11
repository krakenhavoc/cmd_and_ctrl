package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lotus Cobra — Creature — Snake {1}{G}, 2/1 (EDHREC rank 328):
//
//	"Landfall — Whenever a land you control enters, add one mana of
//	 any color."
//
// Every land drop is a Lotus Petal. Landfall is Tireless
// Provisioner's ETB-filtered trigger; the mana is the AddMana
// primitive batch 01 added for Dark Ritual, here with a five-colour
// pipe, so the trigger's resolution queues the same colour pick a
// Birds of Paradise activation does. The mana lands in the pool when
// the trigger RESOLVES — it uses the stack, unlike a mana ability —
// and empties with the pool at the end of the step, so a land played
// in the main phase pays for a spell in that main phase, as printed.
//
// Sandbox simplification: the colour pick is narrowed to the
// controller's commander identity, as every five-colour pipe in the
// catalog is (Treasure, Phyrexian Altar). Printed it is "any color";
// off-identity mana can only ever pay a generic cost, so the
// narrowing removes an option that is almost never taken — weaker,
// declared.
func init() {
	Register(Spec{
		OracleID:     "8ad91f64-ccab-4edc-bd54-b2ee9267d614",
		Name:         "Lotus Cobra",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The landfall mana is limited to your commander's color identity instead of any color."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsLand()
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Lotus Cobra — add one mana of any color (landfall)",
					func(g *game.Game, item *game.StackItem) error {
						return AddMana{Produced: "{W|U|B|R|G}"}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
