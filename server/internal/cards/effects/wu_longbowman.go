package effects

// Wu Longbowman — Creature — Human Soldier Archer {2}{U}, 1/1 (#4062):
//
//	"{T}: This creature deals 1 damage to any target.
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
// A Prodigal Sorcerer with a window.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "68fc55b9-55cc-4caf-a052-b44ddc0b346f",
		Name:         "Wu Longbowman",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{T}: This creature deals 1 damage to any target. Activate only during your turn, before attackers are declared.",
			Cost:      TapCost(),
			Targets:   TargetAny(),
			Purpose:   ForTargets(DamageToTarget(0, 1)),
			Condition: DuringYourTurnBeforeAttackersAreDeclared(),
			Effect:    sourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
