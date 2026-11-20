package effects

// Skyhunter Prowler — Creature — Cat Knight {2}{W}, 1/3 (Fifth Dawn, #5604):
//
//	"Flying, vigilance (This creature can't be blocked except by
//	 creatures with flying or reach, and attacking doesn't cause this
//	 creature to tap.)"
//
// Two printed keywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5e0d13b0-d5d0-4c5f-945e-5cda7d980670",
		Name:            "Skyhunter Prowler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
	})
}
