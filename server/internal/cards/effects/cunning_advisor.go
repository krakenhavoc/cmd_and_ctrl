package effects

// Cunning Advisor — Creature — Human Advisor {3}{B}, 1/1 (#4062):
//
//	"{T}: Target opponent discards a card.
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
// The opponent chooses what to discard (CR 701.9a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b80b6dd3-320b-4304-8073-f74c54e9a8fe",
		Name:         "Cunning Advisor",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{T}: Target opponent discards a card. Activate only during your turn, before attackers are declared.",
			Cost:      TapCost(),
			Targets:   TargetPlayer("target opponent", Opponent()),
			Condition: DuringYourTurnBeforeAttackersAreDeclared(),
			Effect:    targetOpponentDiscards(1),
		}},
	})
}
