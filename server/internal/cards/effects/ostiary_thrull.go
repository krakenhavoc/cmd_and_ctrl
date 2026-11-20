package effects

// Ostiary Thrull — Creature — Thrull {3}{B}, 2/2 (Guildpact, #5599):
//
//	"{W}, {T}: Tap target creature."
//
// Nebelgast Beguiler's tapper: the shared tapChosenPermanent body taps
// the creature named as the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1c99c9d9-3ea8-43bc-96d7-4dea12770b5c",
		Name:         "Ostiary Thrull",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W}, {T}: Tap target creature.",
			Cost:    Plus(ManaCost("{W}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
