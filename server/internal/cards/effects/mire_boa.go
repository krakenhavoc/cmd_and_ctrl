package effects

// Mire Boa — Creature — Snake {1}{G}, 2/1 (Planar Chaos, #5595):
//
//	"Swampwalk (This creature can't be blocked as long as defending player
//	 controls a Swamp.)
//	 {G}: Regenerate this creature. (The next time this creature would be
//	 destroyed this turn, instead tap it, remove it from combat, and heal all
//	 damage on it.)"
//
// Swampwalk is an engine keyword carried as printed data (CR 702.14c);
// regeneration is on the ability's own source (CR 701.19),
// lrwB1RegenerateThisRow.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e9a3a466-df78-49a0-add3-975b94c5cd09",
		Name:            "Mire Boa",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
		Activated:       []ActivatedAbility{lrwB1RegenerateThisRow("{G}")},
	})
}
