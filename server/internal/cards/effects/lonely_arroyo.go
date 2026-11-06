package effects

// Lonely Arroyo — Land — Desert (Outlaws of Thunder Junction, #4468):
//
//	"This land enters tapped.
//	 When this land enters, it deals 1 damage to target opponent.
//	 {T}: Add {W} or {U}."
//
// The Azorius member of Outlaws of Thunder Junction's tapped Desert
// duals, Eroded Canyon's shape exactly, built by b36DesertDual.
//
// No simplification.
func init() {
	Register(b36DesertDual("36508a8a-d1a4-400e-bfda-09436bc4d5d4", "Lonely Arroyo", "W", "U"))
}
