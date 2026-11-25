package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fire Shrine Keeper — Creature — Elemental {R}, 1/1 (Ixalan, #5532):
//
//	"Menace
//	 {7}{R}, {T}, Sacrifice this creature: It deals 3 damage to each of up
//	 to two target creatures."
//
// The sacrifice is a cost, so the Keeper is gone when the ability
// resolves and still deals the damage as the source (CR 608.2h); each
// target still legal takes 3 at once (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "14537df5-e2c8-4906-99ac-742fb628fe08",
		Name:            "Fire Shrine Keeper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{7}{R}, {T}, Sacrifice this creature: It deals 3 damage to each of up to two target creatures.",
			Cost:    effects.Plus(effects.ManaCost("{7}{R}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("up to two target creatures").WithCount(0, 2),
			Purpose: effects.WithAnswers(effects.ForTargets(effects.DamageToTarget(0, 3)), game.AnswerRemove),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(3),
		}},
	})
}
