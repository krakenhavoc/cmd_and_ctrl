package effects

// Halo Hopper — Artifact Creature — Frog {3} 3/2 (March of the Machine):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)"
//
// Convoke is the TapCost, as on every convoke card (CR 702.51).
func init() {
	Register(Spec{
		OracleID:     "f48d30ee-bba6-4449-8c50-caf0d7f06528",
		Name:         "Halo Hopper",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
	})
}
