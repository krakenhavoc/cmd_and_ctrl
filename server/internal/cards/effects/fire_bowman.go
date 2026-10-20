package effects

// Fire Bowman — Creature — Human Soldier Archer {R}, 1/1 (#4062):
//
//	"Sacrifice this creature: It deals 1 damage to any target.
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
// The sacrifice is a cost, so the damage is dealt by the creature as it
// last existed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3f23d7d3-28c6-4388-afb5-8f101b973083",
		Name:         "Fire Bowman",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "Sacrifice this creature: It deals 1 damage to any target. Activate only during your turn, before attackers are declared.",
			Cost:      SacrificeThis(),
			Targets:   TargetAny(),
			Purpose:   ForTargets(DamageToTarget(0, 1)),
			Condition: DuringYourTurnBeforeAttackersAreDeclared(),
			Effect:    b33DamageChosenTargetFromSource(1),
		}},
	})
}
