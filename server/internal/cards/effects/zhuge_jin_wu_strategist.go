package effects

// Zhuge Jin, Wu Strategist — Legendary Creature — Human Advisor {1}{U}{U}, 1/1 (#4062):
//
//	"{T}: Target creature can't be blocked this turn.
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
// The window closes at the declare attackers step, and an evasion granted
// that early is still the whole point: it is in place before blockers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ce995d3f-be7c-4d2a-8ec4-c7f7f831f0bf",
		Name:         "Zhuge Jin, Wu Strategist",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{T}: Target creature can't be blocked this turn. Activate only during your turn, before attackers are declared.",
			Purpose:   ForTargets(BenefitUntilEOT(0, 0, 0)),
			Cost:      TapCost(),
			Targets:   TargetCreature("target creature"),
			Condition: DuringYourTurnBeforeAttackersAreDeclared(),
			Effect:    targetCantBeBlockedThisTurn("Zhuge Jin"),
		}},
	})
}
