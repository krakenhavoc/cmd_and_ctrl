package effects

// Thraben Heretic — Creature — Human Wizard {1}{W}, 2/2 (Dark Ascension, #5571):
//
//	"{T}: Exile target creature card from a graveyard."
//
// Any player's graveyard. A card that left the graveyard in response is
// left alone (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "67247945-17b7-4cb8-ba65-d0dfec367b77",
		Name:         "Thraben Heretic",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Exile target creature card from a graveyard.",
			Cost:    TapCost(),
			Targets: TargetCardInGraveyard("target creature card in a graveyard", Creature()),
			Effect:  exileFirstLegalCardTarget,
		}},
	})
}
