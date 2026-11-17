package effects

// Bola Warrior — Creature — Human Spellshaper Warrior {1}{R}, 1/1
// (Nemesis, #5618):
//
//	"{R}, {T}, Discard a card: Target creature can't block this turn."
//
// The spellshaper cost (nemSpellshaperCost) on Jamuraan Lion's body,
// targetCreatureCantBlockThisTurn: the restriction ends in the cleanup
// step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "13f41253-364f-40a5-a7cd-84d8785b2460",
		Name:         "Bola Warrior",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}, {T}, Discard a card: Target creature can't block this turn.",
			Cost:    nemSpellshaperCost("{R}"),
			Targets: TargetCreature("target creature"),
			Purpose: ForTargets(CantBlockUntilEOT(0, 0, 0)),
			Effect:  targetCreatureCantBlockThisTurn("Bola Warrior — target creature can't block this turn"),
		}},
	})
}
