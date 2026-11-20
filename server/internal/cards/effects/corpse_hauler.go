package effects

// Corpse Hauler — Creature — Human Rogue {1}{B}, 2/1 (Magic 2014, #5563):
//
//	"{2}{B}, Sacrifice this creature: Return target creature card from your graveyard to your hand."
//
// Golgari Guildmage's return with the Hauler itself as the sacrifice. The
// target is chosen before the cost is paid (CR 602.2b), so the Hauler can't
// return itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c2ba1c3b-08c7-4835-81f4-2f98843fb7b8",
		Name:         "Corpse Hauler",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}, Sacrifice this creature: Return target creature card from your graveyard to your hand.",
			Cost:    Plus(ManaCost("{2}{B}"), SacrificeThis()),
			Targets: TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn()),
			Effect:  returnFirstLegalGraveyardTargetToHand,
		}},
	})
}
