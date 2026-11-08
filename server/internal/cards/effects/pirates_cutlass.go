package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pirate's Cutlass — Artifact — Equipment {3} (Foundations, #4465):
//
//	"When this Equipment enters, attach it to target Pirate you control.
//	 Equipped creature gets +2/+1.
//	 Equip {2}"
//
// Dwarven Mattock's shape with a Pirate: a targeted enters trigger that
// resolves through AttachSourceToTarget. With no Pirate to name, the
// trigger leaves the stack and the Cutlass simply stays unattached until
// it is equipped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ac77591c-39d7-4fd1-a6e9-805f57c6614a",
		Name:         "Pirate's Cutlass",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 1)},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Pirate's Cutlass — attach it to target Pirate you control", AttachSourceToTarget),
				PermanentYouControl("target Pirate you control", Creature(), HasSubtype("Pirate")),
			),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
