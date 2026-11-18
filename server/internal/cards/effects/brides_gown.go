package effects

// Bride's Gown — Artifact — Equipment {1}{W} (Innistrad: Crimson Vow,
// #5495):
//
//	"Equipped creature gets +2/+0. It gets an additional +0/+2 and has
//	 first strike as long as an Equipment named Groom's Finery is attached
//	 to a creature you control.
//	 Equip {2}"
//
// The statics are s82VowB06EquipmentPair's, which read the partner Equipment
// by name live, so attaching or unattaching Groom's Finery changes the
// equipped creature at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "63e6c80a-a7c7-4c98-b760-271c36425dd1",
		Name:         "Bride's Gown",
		Completeness: CompletenessFull,
		Static:       s82VowB06EquipmentPair("Groom's Finery", "first strike"),
		Activated:    []ActivatedAbility{EquipAbility("{2}")},
	})
}
