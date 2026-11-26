package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stormrider Rig — Artifact — Equipment {2} (Dragons of Tarkir):
//
//	"Equipped creature gets +1/+1.
//	 Whenever a creature you control enters, you may attach this Equipment to it.
//	 Equip {2}"
//
// The trigger is optional and attaches only if the Rig and the creature are both
// still on the battlefield as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "51c7beca-ca19-4a4d-9512-0255146a9cd0",
		Name:         "Stormrider Rig",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered:    []game.TriggeredAbility{effects.AttachThisToTheEnteringCreature("Stormrider Rig")},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
