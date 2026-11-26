package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Riot Gear — Artifact — Equipment {2} (Gatecrash, #5565):
//
//	"Equipped creature gets +1/+2.
//	 Equip {2}"
//
// A layer 7c static on the host (PumpAttached), moving with the Equipment
// (CR 301.5); equip is the ordinary sorcery-speed activated ability
// (CR 702.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3e150025-51c2-4b92-970c-a89e83210e68",
		Name:         "Riot Gear",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 2)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
