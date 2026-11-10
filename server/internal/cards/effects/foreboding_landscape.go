package effects

// Foreboding Landscape — Land (Tarkir: Dragonstorm Commander):
//
//	"{T}: Add {C}.
//	 {T}, Sacrifice this land: Search your library for a basic Swamp,
//	 Forest, or Island card, put it onto the battlefield tapped, then
//	 shuffle.
//	 Cycling {B}{G}{U}"
//
// The Landscape cycle's shape (rfReprintALandscape) for the Sultai colors.
//
// No simplification.
func init() {
	Register(rfReprintALandscape("bcfe1653-e602-4d38-abe7-bfcc7f203f9d", "Foreboding Landscape",
		"{B}{G}{U}", "Swamp", "Forest", "Island"))
}
