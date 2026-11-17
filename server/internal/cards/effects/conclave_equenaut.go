package effects

// Conclave Equenaut — Creature — Human Soldier {4}{W}{W}, 3/3 (Ravnica:
// City of Guilds, #5600):
//
//	"Convoke
//	 Flying"
//
// Convoke is the spell's tap cost (Spec.TapCost, CR 702.51): each creature
// tapped pays for {1} or one mana of its colour. Flying is the printed
// keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4127794b-84d1-4efb-b0e9-543a02b07fa8",
		Name:            "Conclave Equenaut",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		TapCost:         Convoke(),
	})
}
