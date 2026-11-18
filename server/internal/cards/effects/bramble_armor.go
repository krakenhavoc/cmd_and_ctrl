package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bramble Armor — Artifact — Equipment {1}{G} (Innistrad: Midnight Hunt):
//
//	"When this Equipment enters, attach it to target creature you control.
//	 Equipped creature gets +2/+1.
//	 Equip {4}"
//
// The entry trigger is Baseball Bat's attach (AttachSourceToTarget): it
// targets on the stack, and a target that has gone gets nothing
// (CR 608.2b). The pump is PumpAttached.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "25328cab-c7c0-4c55-99bb-a4149f6c511a",
		Name:         "Bramble Armor",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 1)},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Bramble Armor — attach it to target creature you control", AttachSourceToTarget),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		Activated: []ActivatedAbility{EquipAbility("{4}")},
	})
}
