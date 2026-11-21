package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Malamet Scythe — Artifact — Equipment {2}{G} (The Lost Caverns of
// Ixalan, #4470):
//
//	"Flash
//	 When this Equipment enters, attach it to target creature you control.
//	 Equipped creature gets +2/+2.
//	 Equip {4}"
//
// Biorganic Carapace's enters attach on a flash Equipment (Celestial
// Armor's flash). The trigger targets on the stack; a target that has
// gone gets nothing (CR 608.2b), and a Scythe that has left in response
// attaches to nothing (CR 701.3b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "99c5630e-cc04-4203-9006-3bc6a6807762",
		Name:            "Malamet Scythe",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Malamet Scythe — attach it to target creature you control", effects.AttachSourceToTarget),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{4}")},
	})
}
