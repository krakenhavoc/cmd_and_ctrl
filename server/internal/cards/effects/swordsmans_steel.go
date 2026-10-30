package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Swordsman's Steel — Artifact — Equipment {4} (Marvel Super Heroes
// Commander, #4455):
//
//	"When this Equipment enters, draw a card for each Equipment you
//	 control.
//	 Equipped creature gets +2/+2 for each Equipment you control.
//	 Equip {3}"
//
// Blackblade Reforged's PumpAttachedPer: the count is re-read on every
// layer pass, so the bonus follows the board. "You control" is the
// Equipment's controller (CR 109.5); every Equipment counts, this one and
// unattached ones included. The enters trigger counts as it resolves, so
// an Equipment that arrived or left in between changes the number.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e36100d6-d9ae-454a-bba1-b50a69d5a28f",
		Name:         "Swordsman's Steel",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Swordsman's Steel — draw a card for each Equipment you control",
				func(g *game.Game, item *game.StackItem) error {
					n := countControlled(g, item.Controller, func(c game.Card) bool { return c.HasSubtype("Equipment") })
					return DrawCards{N: n}.Apply(NewContext(g, item))
				}),
		},
		Static: []game.StaticAbility{
			PumpAttachedPer(2, 2, equipmentCountForSteel),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}

// equipmentCountForSteel counts "Equipment you control" for a layer pass.
func equipmentCountForSteel(g *game.Game, source *game.Card) int {
	return permanentsYouControl(g, source, func(c *game.Card) bool { return c.HasSubtype("Equipment") })
}
