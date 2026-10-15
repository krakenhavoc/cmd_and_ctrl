package effects

// Magical Hack — Instant {U}:
//
//	"Change the text of target spell or permanent by replacing all
//	 instances of one basic land type with another. (For example, you
//	 may change "swampwalk" to "plainswalk." This effect lasts
//	 indefinitely.)"
//
// CR 612.2, ADR 0160's 2026-10-15 amendment (text_words.go). A basic
// land type is a word of the type line too ("Alters all occurrences of
// the chosen word in the text box and the type line"), so a Forest
// changed to an Island IS an Island and taps for {U} (CR 305.6), and
// its name stays Forest. Landwalk follows ("swampwalk" becomes
// "plainswalk"), and so does a target clause or a static that names the
// type through its source ("target Swamp", a Swamp lord).
//
// Graded Full for the same reason as Sleight of Mind: the words the
// substitution cannot reach are listed in the amendment as the target's
// gaps.
func init() {
	Register(Spec{
		OracleID:     "cba229fa-9035-405b-b091-3798898a37ee",
		Name:         "Magical Hack",
		Completeness: CompletenessFull,
		Targets:      TargetSpellOrPermanent("target spell or permanent", nil, nil),
		OnResolve:    changeTextOfTarget(textLandWords, false, nil),
	})
}
