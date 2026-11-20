package effects

// Kitchen Imp — Creature — Imp {3}{B}, 2/2 (Modern Horizons 2, #5501):
//
//	"Flying, haste
//	 Madness {B} (If you discard this card, discard it into exile. When
//	 you do, cast it for its madness cost or put it into your
//	 graveyard.)"
//
// Two printed keywords and Muck Drubb's madness (CR 702.35a), which the
// engine grows from the one cost string.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "eaa2ad31-065d-4aa3-a39c-5e491dd59022",
		Name:            "Kitchen Imp",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Madness:         "{B}",
	})
}
