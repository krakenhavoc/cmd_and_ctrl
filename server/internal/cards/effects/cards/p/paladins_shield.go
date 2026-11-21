package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Paladin's Shield — Artifact — Equipment {1}{W} (Adventures in the
// Forgotten Realms, #5499):
//
//	"Flash
//	 When this Equipment enters, attach it to target creature you control.
//	 Equipped creature gets +0/+2.
//	 Equip {3}"
//
// Shredder's Armor's enter trigger (it targets as it goes on the stack, so
// a target that has gone leaves the Equipment unattached, CR 608.2b) on a
// flash Equipment.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dc548b03-ef5f-40df-8311-35d4cf26f3c5",
		Name:            "Paladin's Shield",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static:          []game.StaticAbility{effects.PumpAttached(0, 2)},
		Triggered:       []game.TriggeredAbility{effects.AfrB06EnterAttach("Paladin's Shield")},
		Activated:       []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
