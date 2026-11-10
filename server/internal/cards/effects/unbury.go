package effects

// Unbury — Instant {1}{B} (Lorwyn Eclipsed, #4458):
//
//	"Choose one —
//	 • Return target creature card from your graveyard to your hand.
//	 • Return two target creature cards that share a creature type from
//	   your graveyard to your hand."
//
// The second bullet is two target clauses, the second related to the first
// (SharesACreatureTypeWithTargetCard): the pair must share a creature type
// when chosen and again as the spell resolves, on the cards' creature types
// (a changeling shares with any creature card; an artifact or legendary
// word never counts). Distinct, because "two target" is one instance of the
// word (CR 115.3). If one card has left the graveyard by then, the other is
// still returned as long as it has a creature type the first had as it left
// (the ruling): each card still legal goes to your hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e2928d3c-b64a-443e-a1c1-fd204bb098ba",
		Name:         "Unbury",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Return target creature card from your graveyard to your hand.",
				TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn()),
				eoeABCharmReturn),
			ModeDoing("Return two target creature cards that share a creature type from your graveyard to your hand.",
				Clauses(
					TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn()),
					Distinct(SharesACreatureTypeWithTargetCard(
						TargetCardInGraveyard("another target creature card from your graveyard that shares a creature type with it", Creature(), YouOwn()), 0)),
				),
				eoeABCharmReturn),
		),
	})
}
