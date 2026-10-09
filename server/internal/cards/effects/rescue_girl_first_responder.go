package effects

// Rescue Girl, First Responder — Legendary Creature — Human Cleric
// {2}{W}, 1/3:
//
//	"Flying
//	 {T}: Return another target permanent you control to its owner's
//	 hand. Activate only during your turn."
//
// The target is chosen at activation from your other permanents, and
// the activation is refused outside your own turn (any step). The
// bounce goes through BounceToHand, so a commander asks its owner about
// the command zone. The tap makes it subject to summoning sickness.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "df6fc0b4-333b-4f6c-804c-ba6c8e29c88e",
		Name:            "Rescue Girl, First Responder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:     "{T}: Return another target permanent you control to its owner's hand. Activate only during your turn",
			Cost:      TapCost(),
			Targets:   Another(TargetPermanent("another target permanent you control", YouControl())),
			Condition: DuringYourTurn(),
			Effect:    bounceChosenTarget,
		}},
	})
}
