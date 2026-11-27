package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sai of the Shinobi — Artifact — Equipment {1} (Planechase 2012, S82 OLD, #5569):
//
//	"Equipped creature gets +1/+1.
//	 Whenever a creature you control enters, you may attach this Equipment to it.
//	 Equip {2}"
//
// Stormrider Rig's shape (S82OldB432AttachOnEnterEquipment): the trigger is
// optional and attaches only if the Equipment and the creature are both still on
// the battlefield as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB432AttachOnEnterEquipment("ec57dcb8-127f-4922-a9f9-5227db1751a9", "Sai of the Shinobi", "{2}"))
}
