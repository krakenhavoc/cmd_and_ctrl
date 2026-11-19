package effects

// Epitaph Golem — Artifact Creature — Golem {5}, 3/5 (Innistrad Remastered,
// #5948):
//
//	"{2}: Put target card from your graveyard on the bottom of your library."
//
// Barkform Harvester's ability on a creature: the one target card in the
// activator's graveyard goes to the bottom of their library (a tuck, CR 701.24),
// if it is still there as the ability resolves.
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
			Targets: TargetCardInGraveyard("target card in your graveyard", YouOwn()),
			Effect:  gapB08TuckTheTargetCard,
		}},
	})
}
