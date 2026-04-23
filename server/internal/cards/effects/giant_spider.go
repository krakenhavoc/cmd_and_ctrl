package effects

// Giant Spider — "Reach."
//
// Canonical reach ground-blocker. 2/4 for 3G. The keyword feeds
// CanBlock: Giant Spider can block flying attackers (Serra Angel,
// Vampire Nighthawk).
func init() {
	Register(Spec{
		OracleID:        "e740ce2f-2134-473c-afa1-1b6d2d1e38ef",
		Name:            "Giant Spider",
		PrintedKeywords: []string{"reach"},
	})
}
