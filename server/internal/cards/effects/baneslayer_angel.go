package effects

// Baneslayer Angel — "Flying, first strike, lifelink."
//
// 5/5 for 3WW. Printed text also includes "protection from Demons
// and from Dragons" — protection (CR 702.16) is deferred to S24
// alongside Mind Control per ADR 0014 §11. Ships here without
// the protection clauses; when S24 lands protection, extend this
// Spec with the predicate-based guards at damage / block / target
// / attach hooks.
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
