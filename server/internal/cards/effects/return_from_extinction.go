package effects

// Return from Extinction — Sorcery {1}{B} (Modern Horizons, #5519):
//
//	"Choose one —
//	 • Return target creature card from your graveyard to your hand.
//	 • Return two target creature cards that share a creature type from your
//	 graveyard to your hand."
//
// Unbury's text at sorcery speed: the second bullet's second clause is
// SharesACreatureTypeWithTargetCard over the first, and a changeling card
// shares a type with any creature card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3937a96c-3be1-465d-98b6-1db67d2215e6",
		Name:         "Return from Extinction",
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
