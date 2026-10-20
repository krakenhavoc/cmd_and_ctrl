package effects

// Steam Catapult — Creature — Human Soldier {3}{W}{W}, 2/3 (#4062):
//
//	"{T}: Destroy target tapped creature.
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
// Royal Assassin with a window: the target is tapped as it is chosen and
// again as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b6ea0c70-022a-43e4-aa40-3dc6a8b3b21e",
		Name:         "Steam Catapult",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{T}: Destroy target tapped creature. Activate only during your turn, before attackers are declared.",
			Cost:      TapCost(),
			Targets:   TargetCreature("target tapped creature", tappedPermanent()),
			Condition: DuringYourTurnBeforeAttackersAreDeclared(),
			Effect:    destroyFirstLegalTarget,
		}},
	})
}
