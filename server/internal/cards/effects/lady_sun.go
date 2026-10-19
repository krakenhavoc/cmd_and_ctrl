package effects

// Lady Sun — Legendary Creature — Human Advisor {1}{U}{U}, 1/1 (#4062):
//
//	"{T}: Return Lady Sun and another target creature to their owners' hands.
//	 Activate only during your turn, before attackers are declared."
//
// The window is DuringYourTurnBeforeAttackersAreDeclared, an activation
// condition (CR 602.1b): your own turn, and only until the first declare
// attackers step of the turn begins - the card's ruling reads it from the
// first combat, so a turn with several combat phases does not open it
// again. Not a sorcery-speed gate: it opens your beginning, combat and
// main-phase steps up to that point, with or without an item on the stack,
// and the view and the legal enumerator read the same closure, so the
// ability greys out when the window shuts.
//
// "Another" is object identity (Another), so a second Lady Sun is a legal
// pick but this one is not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dcc363a7-652d-4d2b-9081-3e5735120f21",
		Name:         "Lady Sun",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{T}: Return Lady Sun and another target creature to their owners' hands. Activate only during your turn, before attackers are declared.",
			Cost:      TapCost(),
			Targets:   Another(TargetCreature("another target creature")),
			Condition: DuringYourTurnBeforeAttackersAreDeclared(),
			Effect:    bounceThisAndAnotherTarget,
		}},
	})
}
