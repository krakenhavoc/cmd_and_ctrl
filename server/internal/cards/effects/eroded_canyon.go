package effects

// Eroded Canyon — Land — Desert (EDHREC rank 3848):
//
//	"This land enters tapped.
//	 When this land enters, it deals 1 damage to target opponent.
//	 {T}: Add {U} or {R}."
//
// The Izzet member of Outlaws of Thunder Junction's tapped Desert
// duals — Abraded Bluffs' and Bristling Backwoods' shape exactly,
// built by b36DesertDual.
//
// No simplification.
func init() {
	Register(b36DesertDual("852c6520-d148-4923-a312-05a9af821f24", "Eroded Canyon", "U", "R"))
}
