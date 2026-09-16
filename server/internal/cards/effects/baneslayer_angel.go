package effects

// Baneslayer Angel — "Flying, first strike, lifelink."
//
// 5/5 for 3WW. Printed text also includes "protection from Demons
// and from Dragons". Protection (CR 702.16) is not implemented: S24
// shipped Mind Control without it, and it is tracked in #662. Ships
// here without the protection clauses, declared in Caveats. The
// shape is #662's ADR to decide. The quality is a subtype, so a
// changeling source counts as a Demon and as a Dragon.
//
// Multi-keyword card — the PrintedKeywords slot feeds all three
// as strings into Characteristic.Abilities; combat reads them via
// HasKeyword.
func init() {
	Register(Spec{
		OracleID:        "0e11792b-7fe5-4208-aa0b-e5d09b2b65fe",
		Name:            "Baneslayer Angel",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"Protection from Demons and from Dragons is missing; only flying, first strike and lifelink work."},
		PrintedKeywords: []string{"flying", "first strike", "lifelink"},
	})
}
