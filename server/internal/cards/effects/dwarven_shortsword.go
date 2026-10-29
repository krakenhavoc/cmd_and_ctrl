package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dwarven Shortsword — Artifact — Equipment {3}{W} (The Hobbit, HOB):
//
//	"When this Equipment enters, create a 2/2 red Dwarf creature token,
//	 then attach this Equipment to it.
//	 Equipped creature gets +1/+2.
//	 Equip {2}"
//
// Giant's Amulet's enter trigger without the payment: the card prints its
// own token, so it is read through the Shortsword's text (createTokenAndAttach
// Source, described), and the attach rides the token's creation. If the
// Shortsword has left in response, the Dwarf is made and nothing is attached
// (CR 701.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9cc52afb-9788-46f6-9f9b-347513f8a64f",
		Name:         "Dwarven Shortsword",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 2)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dwarven Shortsword — create a 2/2 red Dwarf creature token, then attach this Equipment to it",
				func(g *game.Game, item *game.StackItem) error {
					return createTokenAndAttachSource(g, item, "2/2 red Dwarf", true)
				}),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
