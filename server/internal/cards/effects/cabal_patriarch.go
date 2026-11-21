package effects

// Cabal Patriarch — Legendary Creature — Human Wizard {3}{B}{B}{B}, 5/5
// (Commander Masters):
//
//	"{2}{B}, Sacrifice a creature: Target creature gets -2/-2 until end of
//	 turn.
//	 {2}{B}, Exile a creature card from your graveyard: Target creature gets
//	 -2/-2 until end of turn."
//
// Plagued Rusalka's sacrifice outlet and Ashnod's graveyard exile cost, each
// paying for the same -2/-2 until end of turn (layer 7c, the target's
// toughness read by state-based actions). Two printed abilities, two rows.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "70d1caee-1c35-4fa3-bbc3-a7dcc226be90",
		Name:         "Cabal Patriarch",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{2}{B}, Sacrifice a creature: Target creature gets -2/-2 until end of turn.",
				Purpose: ForTargets(HarmUntilEOT(0, -2, -2)),
				Cost:    Plus(ManaCost("{2}{B}"), SacrificeACreature()),
				Targets: TargetCreature("target creature"),
				Effect:  boostTargetUntilEOT("Cabal Patriarch — -2/-2", -2, -2),
			},
			{
				Label:   "{2}{B}, Exile a creature card from your graveyard: Target creature gets -2/-2 until end of turn.",
				Purpose: ForTargets(HarmUntilEOT(0, -2, -2)),
				Cost:    Plus(ManaCost("{2}{B}"), ExileFromGraveyard(1, "a creature card", isCreatureCard)),
				Targets: TargetCreature("target creature"),
				Effect:  boostTargetUntilEOT("Cabal Patriarch — -2/-2", -2, -2),
			},
		},
	})
}
