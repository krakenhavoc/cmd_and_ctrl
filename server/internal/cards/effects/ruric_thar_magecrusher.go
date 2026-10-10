package effects

// Ruric Thar, Magecrusher — Legendary Creature — Ogre Warrior
// {5}{G}{G}, 7/7:
//
//	"This spell can't be countered.
//	 Reach, vigilance, trample
//	 Ruric Thar has hexproof as long as they haven't dealt combat
//	 damage yet."
//
// The conditional hexproof is NOT implemented, and the card is weaker
// than printed for it. "Has hexproof until it deals combat damage"
// needs the layer engine to recompute the moment combat damage is
// dealt, which only battlefield moves, counters and a few designations
// do today; a hexproof grant that never turned off after the first hit
// would be stronger than printed, so the grant is left out entirely.
func init() {
	Register(Spec{
		OracleID:        "453b2b3c-7df5-4515-9565-82f603cf451b",
		Name:            "Ruric Thar, Magecrusher",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"Ruric Thar never has hexproof, including before it has dealt combat damage."},
		PrintedKeywords: []string{"reach", "vigilance", "trample"},
		CantBeCountered: true,
	})
}
