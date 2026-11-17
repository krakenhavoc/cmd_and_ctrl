package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Veteran's Powerblade — Artifact — Equipment {3} (The Brothers' War,
// #5482):
//
//	"Equipped creature gets +2/+0.
//	 Equip Soldier {W}
//	 Equip {2}"
//
// Two equip abilities, Dúnedain Blade's shape: the Soldier-only one has the
// tighter target clause and the cheaper cost, the plain {2} one follows.
// The restriction is checked while activating (CR 702.6c); a creature that
// stops being a Soldier keeps the Powerblade.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "da0b4e01-b312-4f37-a034-824ce58db02e",
		Name:         "Veteran's Powerblade",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 0),
		},
		Activated: []ActivatedAbility{
			EquipOnlyAbility("Equip Soldier {W}", "{W}",
				TargetCreature("target Soldier you control", YouControl(), OfCreatureType("Soldier"))),
			EquipAbility("{2}"),
		},
	})
}
