package effects

// Lounge — Land (Ravnica: Clue Edition, #5929):
//
//	"This land enters tapped.
//	 {T}: Add {B} or {G}.
//	 {4}, {T}: Investigate. (Create a Clue token. It's an artifact with
//	 "{2}, Sacrifice this token: Draw a card.")"
//
// One of the set's eight common duals, registered through
// bnsB21ClueTapland (s82_bns_b21_helpers.go).
//
// No simplification.
func init() {
	Register(bnsB21ClueTapland("98a22879-d0b2-441f-bcf6-a67b4552b73c", "Lounge", "B", "G"))
}
