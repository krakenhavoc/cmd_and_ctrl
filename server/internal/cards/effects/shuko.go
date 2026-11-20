package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shuko — Artifact — Equipment {1} (Betrayers of Kamigawa, #5602):
//
//	"Equipped creature gets +1/+0.
//	 Equip {0} ({0}: Attach to target creature you control. Equip only as
//	 a sorcery.)"
//
// The pump is PumpAttached(1, 0) and the equip is the ordinary sorcery-
// speed EquipAbility at {0} (CR 702.6a), with no once-per-turn limit.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8abe0577-8fdb-4e4a-a871-01b21732c961",
		Name:         "Shuko",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 0)},
		Activated:    []ActivatedAbility{EquipAbility("{0}")},
	})
}
