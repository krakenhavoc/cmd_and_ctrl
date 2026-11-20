package effects

// Epitaph Golem — Artifact Creature — Golem {5}, 3/5
// (Shadows over Innistrad, #5543):
//
//	"{2}: Put target card from your graveyard on the bottom of your
//	 library."
//
// Tomb Trawler's ability exactly (trekB27TuckTargetCardToBottom).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e3483dd4-0118-4143-9b55-51078a6f274a",
		Name:         "Epitaph Golem",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}: Put target card from your graveyard on the bottom of your library.",
			Cost:    ManaCost("{2}"),
			Targets: TargetCardInGraveyard("target card from your graveyard", YouOwn()),
			Effect:  trekB27TuckTargetCardToBottom,
		}},
	})
}
