package effects

// Jagged Barrens — Land — Desert (EDHREC rank 3779):
//
//	"This land enters tapped.
//	 When this land enters, it deals 1 damage to target opponent.
//	 {T}: Add {B} or {R}."
//
// The Rakdos member of Outlaws of Thunder Junction's tapped Desert
// duals — Abraded Bluffs' and Bristling Backwoods' shape exactly,
// built by b36DesertDual.
//
// No simplification.
func init() {
	Register(b36DesertDual("64ee02f1-afdb-474b-a893-31538ad7219a", "Jagged Barrens", "B", "R"))
}
