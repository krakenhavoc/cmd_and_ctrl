package effects

// Krosan Constrictor — Creature — Snake {3}{G}, 2/2 (Torment, #5612):
//
//	"Swampwalk (This creature can't be blocked as long as defending
//	 player controls a Swamp.)
//	 {T}: Target black creature gets -2/-0 until end of turn."
//
// Swampwalk is an engine keyword carried as printed data, Bog Raiders'
// (CR 702.14c). The tap ability is Humbling Elder's -2/-0 body on an
// activated row: the colour is read as the ability is activated and
// again as it resolves (CR 608.2b), so a creature that stopped being
// black in response is untouched. Declared as a harm, so it is aimed at
// an opponent's creature; the shrink ends at cleanup (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c9747b5d-0f6d-4a9a-8921-1d5c0be3bd60",
		Name:            "Krosan Constrictor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target black creature gets -2/-0 until end of turn.",
			Purpose: ForTargets(HarmUntilEOT(0, -2, 0)),
			Cost:    TapCost(),
			Targets: TargetCreature("target black creature", OfColor("B")),
			Effect:  boostTargetUntilEOT("Krosan Constrictor — -2/-0 until end of turn", -2, 0),
		}},
	})
}
