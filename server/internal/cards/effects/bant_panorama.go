package effects

// Bant Panorama — Land:
//
//	"{T}: Add {C}.
//	 {1}, {T}, Sacrifice this land: Search your library for a basic
//	 Forest, Plains, or Island card, put it onto the battlefield
//	 tapped, then shuffle."
//
// A row of the Alara Panorama cycle (b43PanoramaFetch in
// naya_panorama.go); the sacrifice is a cost and the search happens on
// resolution.
//
// No simplification.
func init() {
	Register(b43PanoramaFetch(
		"0a1d817d-dce8-4e83-a380-909f7c9eee46", "Bant Panorama",
		"Forest", "Plains", "Island"))
}
