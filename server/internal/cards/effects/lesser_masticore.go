package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lesser Masticore — Artifact Creature — Masticore {2}, 2/2:
//
//	"As an additional cost to cast this spell, discard a card.
//	 {4}: This creature deals 1 damage to target creature.
//	 Persist"
//
// Persist is PrintedKeywords (#2075).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a6a257bd-ee19-4246-b997-5288e6fa41e1",
		Name:            "Lesser Masticore",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordPersist},
		AdditionalCost:  DiscardCost(1),
		Activated: []ActivatedAbility{{
			Label:   "{4}: This creature deals 1 damage to target creature.",
			Cost:    ManaCost("{4}"),
			Targets: TargetCreature("target creature"),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect:  DealDamageToTheTarget(1),
		}},
	})
}
