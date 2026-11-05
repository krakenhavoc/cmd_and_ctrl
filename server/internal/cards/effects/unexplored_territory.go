package effects

// Unexplored Territory — Land (Star Trek, #4973):
//
//	"{T}: Add {C}.
//	 {1}, {T}, Sacrifice this land: Search your library for a basic land
//	 card, put it onto the battlefield tapped, then shuffle."
//
// Shire Terrace's two lines under a Star Trek name, so it shares that
// card's mana ability and fetch (oneTapSacFetchBasicTapped): the searcher
// picks from the basic land cards, the pick enters tapped and the library
// is shuffled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "6e5d4bbd-4999-4acd-8775-25e02d901035",
		Name:          "Unexplored Territory",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{tapForColorless()},
		Activated:     []ActivatedAbility{oneTapSacFetchBasicTapped()},
	})
}
