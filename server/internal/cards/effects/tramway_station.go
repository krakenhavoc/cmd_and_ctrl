package effects

// Tramway Station — Land (Streets of New Capenna, #5491):
//
//	"This land enters tapped.
//	 {T}: Add {B} or {R}.
//	 {2}{B}{R}, {T}, Sacrifice this land: Draw a card."
//
// One of the set's five sacrifice-lands, registered through
// sncCyclingTapland (s82_snc_b03_helpers.go).
//
// No simplification.
func init() {
	Register(sncCyclingTapland("e90e519d-023e-4d19-85ad-9972a76df3ba", "Tramway Station", "B", "R"))
}
