package effects

// Pang Tong, "Young Phoenix" — Legendary Creature — Human Advisor {1}{W}{W}, 1/2 (#4062):
//
//	"{T}: Target creature gets +0/+2 until end of turn.
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
// A toughness-only boost, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6d9dc632-414e-480c-85ef-c2b98a6460ad",
		Name:         "Pang Tong, \"Young Phoenix\"",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{T}: Target creature gets +0/+2 until end of turn. Activate only during your turn, before attackers are declared.",
			Purpose:   ForTargets(BenefitUntilEOT(0, 0, 2)),
			Cost:      TapCost(),
			Targets:   TargetCreature("target creature"),
			Condition: DuringYourTurnBeforeAttackersAreDeclared(),
			Effect:    boostTargetUntilEOT("Pang Tong — +0/+2", 0, 2),
		}},
	})
}
