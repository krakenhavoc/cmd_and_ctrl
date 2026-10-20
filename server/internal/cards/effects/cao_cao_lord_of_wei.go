package effects

// Cao Cao, Lord of Wei — Legendary Creature — Human Soldier {3}{B}{B}, 3/3 (#4062):
//
//	"{T}: Target opponent discards two cards.
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
// The opponent chooses what to discard (CR 701.9a); a hand of one gives up
// the one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "81a98b14-e827-4241-a401-64890def51af",
		Name:         "Cao Cao, Lord of Wei",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{T}: Target opponent discards two cards. Activate only during your turn, before attackers are declared.",
			Cost:      TapCost(),
			Targets:   TargetPlayer("target opponent", Opponent()),
			Condition: DuringYourTurnBeforeAttackersAreDeclared(),
			Effect:    targetOpponentDiscards(2),
		}},
	})
}
