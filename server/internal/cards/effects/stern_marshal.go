package effects

// Stern Marshal — Creature — Human Soldier {2}{W}, 2/2 (#4062):
//
//	"{T}: Target creature gets +2/+2 until end of turn.
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
// Any creature, yours or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "00602959-e9a1-4706-a1cb-70156a6fc713",
		Name:         "Stern Marshal",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{T}: Target creature gets +2/+2 until end of turn. Activate only during your turn, before attackers are declared.",
			Cost:      TapCost(),
			Targets:   TargetCreature("target creature"),
			Condition: DuringYourTurnBeforeAttackersAreDeclared(),
			Effect:    boostTargetUntilEOT("Stern Marshal — +2/+2", 2, 2),
		}},
	})
}
