package effects

// Festering Gulch — Land — Desert (Outlaws of Thunder Junction, #4468):
//
//	"This land enters tapped.
//	 When this land enters, it deals 1 damage to target opponent.
//	 {T}: Add {B} or {G}."
//
// The Golgari member of Outlaws of Thunder Junction's tapped Desert duals,
// Eroded Canyon's shape exactly, built by b36DesertDual.
//
// No simplification.
func init() {
	Register(b36DesertDual("9d3b60af-3e38-4d36-95fc-11b31c38f955", "Festering Gulch", "B", "G"))
}
