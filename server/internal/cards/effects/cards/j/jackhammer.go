package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jackhammer — Artifact — Equipment {1}{R} (Streets of New Capenna, #5491):
//
//	"Equipped creature gets +2/+0.
//	 Equip {2}"
//
// A plain pump Equipment: PumpAttached and the equip ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f73d89ec-05c9-4998-a24b-bc675b04ec16",
		Name:         "Jackhammer",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 0)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
