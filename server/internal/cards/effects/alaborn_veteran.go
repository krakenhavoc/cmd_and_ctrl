package effects

// Alaborn Veteran — Creature — Human Knight {2}{W}, 2/2 (#4062):
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
// Any creature, yours or not: the card says only "target creature".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7d9e905b-5f5c-4dbb-95cb-ccba0a5cbf8d",
		Name:         "Alaborn Veteran",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{T}: Target creature gets +2/+2 until end of turn. Activate only during your turn, before attackers are declared.",
			Cost:      TapCost(),
			Targets:   TargetCreature("target creature"),
			Condition: DuringYourTurnBeforeAttackersAreDeclared(),
			Effect:    boostTargetUntilEOT("Alaborn Veteran — +2/+2", 2, 2),
		}},
	})
}
