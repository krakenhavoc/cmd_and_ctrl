package effects

// Chainwhip Cyclops — Creature — Cyclops Warrior {4}{R}, 4/4 (Ravnica:
// Clue Edition):
//
//	"{3}{R}: Target creature can't block this turn."
//
// Bola Warrior's body (targetCreatureCantBlockThisTurn) on a plain mana
// cost: the restriction ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a0324f33-3f25-4374-affa-c7b5b54bc8d7",
		Name:         "Chainwhip Cyclops",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}{R}: Target creature can't block this turn.",
			Cost:    ManaCost("{3}{R}"),
			Targets: TargetCreature("target creature"),
			Purpose: ForTargets(CantBlockUntilEOT(0, 0, 0)),
			Effect:  targetCreatureCantBlockThisTurn("Chainwhip Cyclops — target creature can't block this turn"),
		}},
	})
}
