package effects

// Xun Yu, Wei Advisor — Legendary Creature — Human Advisor {1}{B}{B}, 1/1 (#4062):
//
//	"{T}: Target creature you control gets +2/+0 until end of turn.
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
// Only a creature you control, itself included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fb5fea4a-1b04-45c6-a207-e2e8493ea968",
		Name:         "Xun Yu, Wei Advisor",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{T}: Target creature you control gets +2/+0 until end of turn. Activate only during your turn, before attackers are declared.",
			Purpose:   ForTargets(BenefitUntilEOT(0, 2, 0)),
			Cost:      TapCost(),
			Targets:   TargetCreature("target creature you control", YouControl()),
			Condition: DuringYourTurnBeforeAttackersAreDeclared(),
			Effect:    boostTargetUntilEOT("Xun Yu — +2/+0", 2, 0),
		}},
	})
}
