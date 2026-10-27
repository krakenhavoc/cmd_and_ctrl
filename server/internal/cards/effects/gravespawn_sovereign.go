package effects

func init() {
	cost := TapAnUntapped("five untapped Zombies you control", HasSubtype("Zombie"))
	cost.TapOthers.Count = 5
	Register(Spec{OracleID: "4235c99b-50de-4190-8938-3ce5ec5c5248", Name: "Gravespawn Sovereign", Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "Tap five untapped Zombies you control: Put target creature card from a graveyard onto the battlefield under your control.",
			Cost:  cost, Targets: TargetCardInGraveyard("target creature card from a graveyard", Creature()),
			Effect: returnFirstLegalGraveyardTargetUnderYourControl,
		}},
	})
}
