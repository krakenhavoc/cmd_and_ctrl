package effects

// Skyshooter — Creature — Centaur Archer {1}{G}, 1/2 (Odyssey, #5613):
//
//	"Reach (This creature can block creatures with flying.)
//	 {T}, Sacrifice this creature: Destroy target attacking or blocking
//	 creature with flying."
//
// Reach is a printed keyword (CR 702.17). The target must be in combat
// and have flying, both judged at announce and again as the ability
// resolves (CR 608.2b), so a flier removed from combat or stripped of
// flying in response survives. The sacrifice is part of the cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9c5a3407-e3e6-4f5f-9c24-6f24c19e1a2e",
		Name:            "Skyshooter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice this creature: Destroy target attacking or blocking creature with flying.",
			Cost:    Plus(TapCost(), SacrificeThis()),
			Targets: TargetCreature("target attacking or blocking creature with flying", AttackingOrBlocking(), HasKeyword("flying")),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
