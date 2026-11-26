package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slab Hammer — Artifact — Equipment {2} (Battle for Zendikar, #5547):
//
//	"Whenever equipped creature attacks, you may return a land you control to
//	 its owner's hand. If you do, the creature gets +2/+2 until end of turn.
//	 Equip {2}"
//
// The land is an optional pick at resolution; the pump needs the land to
// have really left the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f353948c-b5f9-4840-a89a-19206dde9d47",
		Name:         "Slab Hammer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverEnchantedCreatureAttacks("Slab Hammer — you may return a land you control to its owner's hand; if you do, the creature gets +2/+2 until end of turn", effects.S82OldB314SlabHammer),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
