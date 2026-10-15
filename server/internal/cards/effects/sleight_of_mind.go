package effects

// Sleight of Mind — Instant {U}:
//
//	"Change the text of target spell or permanent by replacing all
//	 instances of one color word with another. (For example, you may
//	 change "target black spell" to "target blue spell." This effect
//	 lasts indefinitely.)"
//
// CR 612.2, ADR 0160's 2026-10-15 amendment. The two words are chosen
// as it resolves (text_words.go), and the change is a layer-3 record
// on the target that every colour reader consults through it: a target
// clause's colour (Doom Blade's "nonblack"), a printed protection or
// hexproof from a colour, and the colour statics that read their source.
// A change made to a permanent spell stays with the permanent.
//
// Graded Full: every word the substitution reaches is changed, and a
// word it cannot reach (a colour hard-coded in some card's own closure,
// listed in the amendment) is the TARGET's gap, not this card's.
func init() {
	Register(Spec{
		OracleID:     "99dba614-40d3-41c1-a3b2-edc8777b010f",
		Name:         "Sleight of Mind",
		Completeness: CompletenessFull,
		Targets:      TargetSpellOrPermanent("target spell or permanent", nil, nil),
		OnResolve:    changeTextOfTarget(textColorWords, false, nil),
	})
}
