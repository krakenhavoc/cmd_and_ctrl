package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Groom's Finery — Artifact — Equipment {1}{B} (Innistrad: Crimson Vow,
// #5495):
//
//	"Equipped creature gets +2/+0. It gets an additional +0/+2 and has
//	 deathtouch as long as an Equipment named Bride's Gown is attached to a
//	 creature you control.
//	 Equip {2}"
//
// The statics are s82VowB06EquipmentPair's, which read the partner Equipment
// by name live, so attaching or unattaching Bride's Gown changes the
// equipped creature at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b7d2feda-1a0e-4b62-be6d-a8dec7d8ea31",
		Name:         "Groom's Finery",
		Completeness: effects.CompletenessFull,
		Static:       effects.S82VowB06EquipmentPair("Bride's Gown", "deathtouch"),
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
