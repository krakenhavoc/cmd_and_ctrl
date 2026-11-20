package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "a6a257bd-ee19-4246-b997-5288e6fa41e1",
		Name:            "Lesser Masticore",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordPersist},
		AdditionalCost:  effects.DiscardCost(1),
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}: This creature deals 1 damage to target creature.",
			Cost:    effects.ManaCost("{4}"),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.DealDamageToTheTarget(1),
		}},
	})
}
