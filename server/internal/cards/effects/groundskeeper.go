package effects

// Groundskeeper — Creature — Human Druid {G}, 1/1 (Foundations
// Commander):
//
//	"{1}{G}: Return target basic land card from your graveyard to your
//	 hand."
//
// Carnivorous Cultivator's graveyard-to-hand return with a basic-land
// target clause on an activated ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f2c19fba-9b06-461f-bb5e-e827ba41204d",
		Name:         "Groundskeeper",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}: Return target basic land card from your graveyard to your hand.",
			Cost:    ManaCost("{1}{G}"),
			Targets: TargetCardInGraveyard("target basic land card from your graveyard", cmdrB9BasicLandCard(), YouOwn()),
			Effect:  returnTargetGraveyardCardToHand,
		}},
	})
}
