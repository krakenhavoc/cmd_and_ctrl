package effects

// Stone Catapult — Creature — Human Soldier {4}{B}, 1/2 (#4062):
//
//	"{T}: Destroy target tapped nonblack creature.
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
// Royal Assassin with a window and a colour bar, both judged at announce
// and again at resolution (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c3ead850-3f22-4edb-ac96-40abbdf27905",
		Name:         "Stone Catapult",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{T}: Destroy target tapped nonblack creature. Activate only during your turn, before attackers are declared.",
			Cost:      TapCost(),
			Targets:   TargetCreature("target tapped nonblack creature", And(tappedPermanent(), NonBlack())),
			Condition: DuringYourTurnBeforeAttackersAreDeclared(),
			Effect:    destroyFirstLegalTarget,
		}},
	})
}
