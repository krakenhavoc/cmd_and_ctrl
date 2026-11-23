package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bronze Sword — Artifact — Equipment {1} (Theros Beyond Death, #5514):
//
//	"Equipped creature gets +2/+0.
//	 Equip {3} ({3}: Attach to target creature you control. Equip
//	 only as a sorcery.)"
//
// Darksteel Axe's pump and equip without its indestructible.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "14691783-1246-45ed-a439-c86d7b5eef62",
		Name:         "Bronze Sword",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 0)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
