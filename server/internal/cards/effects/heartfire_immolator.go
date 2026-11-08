package effects

// Heartfire Immolator — Creature — Human Wizard {1}{R}, 2/2 (Foundations,
// #4465):
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)
//	 {R}, Sacrifice this creature: It deals damage equal to its power to
//	 target creature or planeswalker."
//
// Prowess is the engine's keyword (game/prowess.go). The ability is
// Hellhole Flailer's (sacrificedSourceDealsItsPowerToTheTarget): the
// sacrifice is a cost, so "its power" is its power as it last existed on
// the battlefield (CR 113.7a), prowess boosts included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1e414a4f-6e76-4cbe-9c69-d52852cd5bc0",
		Name:            "Heartfire Immolator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Activated: []ActivatedAbility{{
			Label:   "{R}, Sacrifice this creature: It deals damage equal to its power to target creature or planeswalker.",
			Cost:    Plus(ManaCost("{R}"), SacrificeThis()),
			Targets: TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
			Effect:  sacrificedSourceDealsItsPowerToTheTarget,
		}},
	})
}
