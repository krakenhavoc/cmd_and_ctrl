package effects

// North Pole Patrol — Creature — Human Soldier Ally {2}{U}, 2/3:
//
//	"{T}: Untap another target permanent you control.
//	 Waterbend {3}, {T}: Tap target creature an opponent controls. (While
//	 paying a waterbend cost, you can tap your artifacts and creatures to
//	 help. Each one pays for {1}.)"
//
// The first ability is Kelpie Guide's untap. The second puts the
// waterbend {3} beside the {T} symbol: the Patrol taps itself for the
// {T} (so it can't also be one of the waterbenders, and it needs to have
// been under your control since your turn began, CR 302.6) while other
// untapped artifacts and creatures you control may each pay {1} of the
// {3} (CR 701.67a).
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "a1230e78-0bfa-489e-a1b9-3c227bceb6fb",
		Name:         "North Pole Patrol",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: Untap another target permanent you control.",
				Cost:    TapCost(),
				Targets: Another(TargetPermanent("another target permanent you control", YouControl())),
				Effect:  b22UntapEachLegalTarget,
			},
			{
				Label:   "Waterbend {3}, {T}: Tap target creature an opponent controls.",
				Cost:    Plus(WaterbendCost("{3}"), TapCost()),
				Targets: TargetCreature("target creature an opponent controls", OpponentControls()),
				Effect:  b36TapChosenCreature,
			},
		},
	})
}
