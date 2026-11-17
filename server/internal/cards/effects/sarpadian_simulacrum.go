package effects

// Sarpadian Simulacrum — Artifact Creature — Goblin {R}, 1/1 (Modern
// Horizons 3, #5467):
//
//	"Haste
//	 {3}{R}, Sacrifice this creature: It deals 4 damage to target
//	 creature."
//
// Divebomber Griffin's sacrifice shot: the Simulacrum is sacrificed as
// part of the cost (CR 602.2b) and is still the source of the damage, as
// it last existed on the battlefield (CR 113.7a). A target gone by
// resolution takes nothing (CR 608.2b). Haste is the engine's keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b563a578-86ce-41a5-a83e-c70c6a4960e2",
		Name:            "Sarpadian Simulacrum",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{{
			Label:   "{3}{R}, Sacrifice this creature: It deals 4 damage to target creature.",
			Cost:    Plus(ManaCost("{3}{R}"), SacrificeThis()),
			Targets: TargetCreature("target creature"),
			Purpose: ForTargets(DamageToTarget(0, 4)),
			Effect:  b33DamageChosenTargetFromSource(4),
		}},
	})
}
