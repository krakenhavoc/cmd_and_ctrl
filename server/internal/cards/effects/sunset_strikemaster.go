package effects

// Sunset Strikemaster — Creature — Human Monk {1}{R}, 3/1 (Tarkir:
// Dragonstorm, #4463):
//
//	"{T}: Add {R}.
//	 {2}{R}, {T}, Sacrifice this creature: It deals 6 damage to target
//	 creature with flying."
//
// A plain {T}: Add {R} mana ability, and Aerie Ouphes' flying-only target
// on a Caustic Caterpillar-style self-sacrifice. The Strikemaster is
// sacrificed as the ability is activated; the ability still resolves
// without it, the damage dealt by the Strikemaster as it last existed, and
// the target is checked again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d81ade13-e7c7-4fe3-bb0c-5007ed9f1198",
		Name:         "Sunset Strikemaster",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "{T}: Add {R}.",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}, {T}, Sacrifice this creature: It deals 6 damage to target creature with flying.",
			Cost:    Plus(ManaCost("{2}{R}"), TapCost(), SacrificeThis()),
			Targets: TargetCreature("target creature with flying", HasKeyword("flying")),
			Purpose: ForTargets(DamageToTarget(0, 6)),
			Effect:  sourceDealsDamageToEachLegalTarget(6),
		}},
	})
}
