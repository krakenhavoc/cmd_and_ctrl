package effects

// Botanical Plaza — Land (Streets of New Capenna, #5491):
//
//	"This land enters tapped.
//	 {T}: Add {G} or {W}.
//	 {2}{G}{W}, {T}, Sacrifice this land: Draw a card."
//
// One of the set's five sacrifice-lands, registered through
// sncCyclingTapland (s82_snc_b03_helpers.go).
//
// No simplification.
func init() {
	Register(sncCyclingTapland("19e6ebaf-7148-428f-8c79-1e9e57a4e659", "Botanical Plaza", "G", "W"))
}
