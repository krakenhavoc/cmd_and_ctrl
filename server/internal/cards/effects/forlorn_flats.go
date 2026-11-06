package effects

// Forlorn Flats — Land — Desert (Outlaws of Thunder Junction, #4468):
//
//	"This land enters tapped.
//	 When this land enters, it deals 1 damage to target opponent.
//	 {T}: Add {W} or {B}."
//
// The Orzhov member of Outlaws of Thunder Junction's tapped Desert duals,
// Eroded Canyon's shape exactly, built by b36DesertDual.
//
// No simplification.
func init() {
	Register(b36DesertDual("ebb3e2ff-2214-4e11-88fb-e0fa84288cf1", "Forlorn Flats", "W", "B"))
}
