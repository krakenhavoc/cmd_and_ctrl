package effects

// Conservatory — Land (Ravnica: Clue Edition, #5929):
//
//	"This land enters tapped.
//	 {T}: Add {G} or {W}.
//	 {4}, {T}: Investigate. (Create a Clue token. It's an artifact with
//	 "{2}, Sacrifice this token: Draw a card.")"
//
// One of the set's eight common duals, registered through
// bnsB21ClueTapland (s82_bns_b21_helpers.go).
//
// No simplification.
func init() {
	Register(bnsB21ClueTapland("8f88b0bf-81bd-4223-9dad-d8c49ff4b87a", "Conservatory", "G", "W"))
}
