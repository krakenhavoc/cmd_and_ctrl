package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Greataxe — Artifact — Equipment {1} (Adventures in the Forgotten
// Realms, #5499):
//
//	Equipped creature gets +4/+0.
//	Equip {5}
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "843f3373-b8ef-416c-9163-aa1aa6b4a51d",
		Name:         "Greataxe",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(4, 0)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{5}")},
	})
}
