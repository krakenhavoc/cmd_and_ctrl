package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Warmonger's Chariot — Artifact — Equipment {2} (Jumpstart, #5510):
//
//	"Equipped creature gets +2/+2.
//	 As long as equipped creature has defender, it can attack as though it
//	 didn't have defender.
//	 Equip {3}"
//
// Behemoth Sledge's pump and equip. The waiver is a layer-6 "can attack as
// though it didn't have defender" on the equipped creature: it changes
// nothing for a creature without defender, so reading it unconditionally is
// the printed condition. The keyword stays.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "90f4ec87-d33d-46c1-8cc7-3e7ad3f8ad70",
		Name:         "Warmonger's Chariot",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.EquippedCanAttackAsThoughNoDefender("As long as equipped creature has defender, it can attack as though it didn't have defender."),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
