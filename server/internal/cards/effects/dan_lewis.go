package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dan Lewis — Legendary Creature — Human {1}{R}, 2/2 (Doctor Who):
//
//	"Noncreature, non-Equipment artifacts you control are Equipment in addition to
//	 their other types and have "Equipped creature gets +1/+0" and equip {1}.
//	 Doctor's companion (You can have two commanders if the other is the Doctor.)"
//
// Gemcutter Buccaneer's shape (granted_equipment.go): a layer-4 static adds the
// Equipment type to each matching artifact, and layer-6 grants give it the
// "Equipped creature gets +1/+0" static and equip {1}. The set is read each time, so
// an artifact that stops being a creature, or that becomes an Equipment on its own,
// enters or leaves it at once; the grants keep to the set the type addition began
// with (CR 613.6), judging "non-Equipment" on the printed subtypes. An artifact creature (a Vehicle
// that is crewed) is not matched. Doctor's companion is a deck-construction rule
// read off the oracle text.
//
// No simplification.
func init() {
	const (
		pumpGrant  = "dan-lewis/artifact-pump"
		equipGrant = "dan-lewis/artifact-equip"
	)
	pump := PumpAttached(1, 0)
	pump.Label = "Equipped creature gets +1/+0."
	Register(Spec{
		OracleID:     "b84ab039-901e-4635-b57e-3f2dac239f38",
		Name:         "Dan Lewis",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{Key: pumpGrant, Static: []game.StaticAbility{pump}, Text: "Equipped creature gets +1/+0."},
			{Key: equipGrant, Activated: []ActivatedAbility{EquipAbility("{1}")}, Text: "Equip {1}"},
		},
		Static: []game.StaticAbility{
			AreAlsoEquipment("Noncreature, non-Equipment artifacts you control are Equipment in addition to their other types.", gapB87NoncreatureNonEquipmentArtifactsYouControl),
			Labelled(GrantAbilities(gapB87DanLewisGrantSet, pumpGrant, equipGrant),
				"Noncreature, non-Equipment artifacts you control have \"Equipped creature gets +1/+0\" and \"Equip {1}\"."),
		},
	})
}
