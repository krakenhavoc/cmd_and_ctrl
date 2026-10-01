package effects

// Flailing Manticore — Creature — Manticore {3}{R}, 3/3:
//
//	"Flying, first strike
//	 {1}: This creature gets +1/+1 until end of turn. Any player may
//	 activate this ability.
//	 {1}: This creature gets -1/-1 until end of turn. Any player may
//	 activate this ability."
//
// Flying and first strike are printed keywords (CR 702.9, 702.7). The
// two activations are the Flailing pair (flailing_helpers.go): any
// player may activate either (CR 602.2, 602.1b) and pays its {1} out of
// their own pool (CR 602.1a), and the Manticore itself grows or shrinks
// until end of turn (CR 611.2c, 514.2). A -1/-1 that takes its
// toughness to 0 puts it into its owner's graveyard at the next
// state-based check (CR 704.5f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "044a169e-32ae-464a-a642-8d3409e82d9f",
		Name:            "Flailing Manticore",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
		Activated:       flailingAbilities("Flailing Manticore"),
	})
}
