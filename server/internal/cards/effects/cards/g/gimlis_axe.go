package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gimli's Axe — Artifact — Equipment {2}{R} (The Lord of the Rings: Tales of
// Middle-earth):
//
//	"Equipped creature gets +3/+0.
//	 As long as equipped creature is legendary, it has menace.
//	 Equip {2}"
//
// The pump is PumpAttached. The conditional menace is Bilbo's Ring's shape
// (GrantToAttachedWhile): read live in layer 6 over the host as the earlier
// layers left it, so a creature that is or stops being legendary gains or
// loses menace with no further event.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "773b40b1-72f0-48a0-b210-2afa5df5485e",
		Name:         "Gimli's Axe",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 0),
			effects.GrantToAttachedWhile(func(host *game.Card, _ *game.Game, _ *game.Card) bool { return host.IsLegendary() }, "menace"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
