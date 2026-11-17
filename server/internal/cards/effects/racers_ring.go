package effects

// Racers' Ring — Land (Streets of New Capenna, #5491):
//
//	"This land enters tapped.
//	 {T}: Add {R} or {G}.
//	 {2}{R}{G}, {T}, Sacrifice this land: Draw a card."
//
// One of the set's five sacrifice-lands, registered through
// sncCyclingTapland (s82_snc_b03_helpers.go).
//
// No simplification.
func init() {
	Register(sncCyclingTapland("7e2eb4d5-22a3-43c1-8cf3-e85723da1b61", "Racers' Ring", "R", "G"))
}
