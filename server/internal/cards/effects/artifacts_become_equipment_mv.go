package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// artifacts_become_equipment_mv.go — #6266: Bludgeon Brawl's "Each
// noncreature, non-Equipment artifact is an Equipment with equip {X} and
// 'Equipped creature gets +X/+0,' where X is that artifact's mana value."
//
// Dan Lewis's shape (granted_equipment.go, AreAlsoEquipment plus layer-6
// grants), widened to every player's artifacts and with X read off each
// artifact: the equip row's cost is game.AbilityCost.ManaValueFromSource
// (#5429), which the granted-row builder writes as "{N}" for the artifact
// that has the row (CR 202.3), and the pump reads the mana value of the
// Equipment it is on, continuously. Equip is CR 702.6a's ability.

// NoncreatureNonEquipmentArtifacts is "each noncreature, non-Equipment
// artifact", every player's, read post-layer (the layer-4 type addition's
// set).
func NoncreatureNonEquipmentArtifacts(target *game.Card, _ *game.Game, _ *game.Card) bool {
	return target.HasCardType("artifact") && !target.IsCreature() && !target.HasSubtype("Equipment")
}

// NoncreatureNonEquipmentArtifactsGrantSet is the same set for the
// layer-6 grants. One effect applying in several layers keeps to the set
// it began with (CR 613.6), so an artifact this effect made an Equipment
// in layer 4 still gets the grants: "non-Equipment" is judged on its
// printed subtypes, as Dan Lewis's grant set judges it.
func NoncreatureNonEquipmentArtifactsGrantSet(target *game.Card, _ *game.Game, _ *game.Card) bool {
	if !target.HasCardType("artifact") || target.IsCreature() {
		return false
	}
	_, _, printed := game.ParseTypeLine(target.TypeLine)
	for _, st := range printed {
		if st == "Equipment" {
			return false
		}
	}
	return true
}

// PumpAttachedBySourceManaValue is "Equipped creature gets +X/+0, where X
// is that artifact's mana value": a layer 7c modify (CR 613.4c) by the
// mana value of the Equipment it is on, read as the layers are applied.
func PumpAttachedBySourceManaValue(label string) game.StaticAbility {
	return game.StaticAbility{
		Label:     label,
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			if source != nil {
				c.Power += source.ManaValue()
			}
		},
	}
}

// EquipForItsManaValue is "equip {X}, where X is this artifact's mana
// value" (CR 702.6a): EquipAbility whose mana component is the mana
// value of the permanent that has the row. Only in a granted bundle (the
// row builder writes the number; game.AbilityCost.ManaValueFromSource).
func EquipForItsManaValue() ActivatedAbility {
	ab := EquipAbility("{X}")
	ab.Cost = ItsManaValue()
	return ab
}
