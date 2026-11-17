package effects

// Autochthon Wurm — Creature — Wurm {10}{G}{G}{G}{W}{W}, 9/14 (Ravnica:
// City of Guilds, #5600):
//
//	"Convoke
//	 Trample"
//
// Convoke is the spell's tap cost (Spec.TapCost, CR 702.51), which is how
// fifteen mana gets paid: each creature tapped pays for {1} or one mana of
// its colour, coloured pips included. Trample is the printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "18379907-ddba-479e-8953-3adc1ed40458",
		Name:            "Autochthon Wurm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		TapCost:         Convoke(),
	})
}
