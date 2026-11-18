package effects

// Lambholt Harrier — Creature — Wolf {1}{R}, 2/2 (Innistrad: Midnight Hunt, S82):
//
//	"{3}{R}: Target creature can't block this turn."
//
// Bola Warrior's restriction (targetCreatureCantBlockThisTurn), paid with mana only.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "84c7d60a-7a7a-4344-8db5-d658792e3ac4",
		Name:         "Lambholt Harrier",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}{R}: Target creature can't block this turn.",
			Cost:    ManaCost("{3}{R}"),
			Targets: TargetCreature("target creature"),
			Purpose: ForTargets(CantBlockUntilEOT(0, 0, 0)),
			Effect:  targetCreatureCantBlockThisTurn("Lambholt Harrier — target creature can't block this turn"),
		}},
	})
}
