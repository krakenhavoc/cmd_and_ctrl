package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "dc548b03-ef5f-40df-8311-35d4cf26f3c5",
		Name:            "Paladin's Shield",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static:          []game.StaticAbility{PumpAttached(0, 2)},
		Triggered:       []game.TriggeredAbility{afrB06EnterAttach("Paladin's Shield")},
		Activated:       []ActivatedAbility{EquipAbility("{3}")},
	})
}
