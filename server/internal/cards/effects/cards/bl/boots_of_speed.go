package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boots of Speed — Artifact — Equipment {R} (Adventures in the Forgotten
// Realms, #5499):
//
//	Equipped creature gets +1/+0 and has haste.
//	Equip {1}
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f6601bc-cb12-4054-9027-e4a0ebdae1a3",
		Name:         "Boots of Speed",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 0), effects.GrantToAttached("haste")},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
