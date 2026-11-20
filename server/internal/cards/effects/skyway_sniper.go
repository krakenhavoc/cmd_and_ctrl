package effects

// Skyway Sniper — Creature — Elf Archer {G}, 1/2 (Core Set 2021, #5511):
//
//	"Reach
//	 {2}{G}: This creature deals 1 damage to target creature with flying."
//
// Centaur Archer's shot at a flyer (DealDamageToTheTarget) for a mana
// cost rather than a tap; the flying clause is re-checked as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8d907825-25f9-4b25-9fac-e0a207ed1797",
		Name:            "Skyway Sniper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}: This creature deals 1 damage to target creature with flying.",
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Cost:    ManaCost("{2}{G}"),
			Targets: TargetCreature("target creature with flying", HasKeyword("flying")),
			Effect:  DealDamageToTheTarget(1),
		}},
	})
}
