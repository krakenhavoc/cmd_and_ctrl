package effects

// Vampire Nighthawk — "Flying, deathtouch, lifelink."
//
// Triple-threat creature — 2/3 for 1BB. The three keywords all
// flow through the same PrintedKeywords slot. Combat interactions
// exercise deathtouch (flag any blocker for destruction) + lifelink
// (source controller gains damage as life) + flying (only flying /
// reach blockers can stop it).
func init() {
	Register(Spec{
		OracleID:        "feb244f8-bcb1-44cf-9940-2719221a7309",
		Name:            "Vampire Nighthawk",
		PrintedKeywords: []string{"flying", "deathtouch", "lifelink"},
	})
}
