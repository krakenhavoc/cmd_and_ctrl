package effects

// Royal Assassin — Creature — Human Assassin {1}{B}{B}, 1/1 (EDHREC
// rank 2258):
//
//	"{T}: Destroy target tapped creature."
//
// The original repeatable removal. A CR 602 activation with a tap
// cost — summoning sickness applies (CR 302.1), the engine enforces
// it — targeting a creature that is tapped RIGHT NOW: the Wandering
// Emperor's tappedPermanent predicate keeps an untapped creature out
// of the picker, and CR 608.2b re-checks it on resolution, so a
// creature untapped in response is a fizzle. An attacking creature is
// tapped by its declaration and is the printed prey; a vigilance
// attacker is not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9ed6f28f-a3db-48c5-9ab0-b90a7fba5f57",
		Name:         "Royal Assassin",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Destroy target tapped creature.",
			Cost:    TapCost(),
			Targets: TargetCreature("target tapped creature", tappedPermanent()),
			Effect:  b17DestroyFirstLegalTarget,
		}},
	})
}
