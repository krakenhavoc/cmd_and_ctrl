package effects

// Esper Panorama — Land:
//
//	"{T}: Add {C}.
//	 {1}, {T}, Sacrifice this land: Search your library for a basic
//	 Plains, Island, or Swamp card, put it onto the battlefield tapped,
//	 then shuffle."
//
// A row of the Panorama cycle (b43PanoramaFetch, see Naya Panorama).
//
// No simplification.
func init() {
	Register(b43PanoramaFetch(
		"6b9cd3d0-4316-4945-b960-12f51052d260", "Esper Panorama",
		"Plains", "Island", "Swamp"))
}
