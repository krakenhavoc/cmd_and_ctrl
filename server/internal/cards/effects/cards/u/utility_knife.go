package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Utility Knife — Artifact — Equipment {1} (Zendikar Rising, #5508):
//
//	"When this Equipment enters, attach it to target creature you control.
//	 Equipped creature gets +1/+1.
//	 Equip {3} ({3}: Attach to target creature you control. Equip only as a
//	 sorcery.)"
//
// Baseball Bat's shape: the enters trigger attaches it to the target chosen as
// it goes on the stack, the pump is a layer 7c static on the attached
// creature, and equip is the standard sorcery-speed ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b546c78c-5180-46c3-8594-b9d5ad9c3988",
		Name:         "Utility Knife",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered:    []game.TriggeredAbility{effects.AfrB06EnterAttach("Utility Knife")},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
