package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "99c5630e-cc04-4203-9006-3bc6a6807762",
		Name:            "Malamet Scythe",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Malamet Scythe — attach it to target creature you control", AttachSourceToTarget),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		Activated: []ActivatedAbility{EquipAbility("{4}")},
	})
}
