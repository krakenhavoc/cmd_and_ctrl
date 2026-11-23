package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "da0b4e01-b312-4f37-a034-824ce58db02e",
		Name:         "Veteran's Powerblade",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 0),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipOnlyAbility("Equip Soldier {W}", "{W}",
				effects.TargetCreature("target Soldier you control", effects.YouControl(), effects.OfCreatureType("Soldier"))),
			effects.EquipAbility("{2}"),
		},
	})
}
