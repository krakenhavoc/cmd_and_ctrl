package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barrels of Blasting Jelly — Artifact {1} (Avatar: The Last Airbender,
// #4459):
//
//	"{1}: Add one mana of any color. Activate only once each turn.
//	 {5}, {T}, Sacrifice this artifact: It deals 5 damage to target
//	 creature."
//
// The filter is a mana ability with the engine's OncePerTurn bit (Mana
// Bloom's), so every path refuses a second activation in a turn. The
// second ability is Disruptor Pistol's at a different cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6cd520c-106b-4ee5-bc02-6c5fbdbc78dd",
		Name:         "Barrels of Blasting Jelly",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:        effects.ManaAbilityCost{Mana: "{1}"},
			Produced:    "{W|U|B|R|G}",
			Label:       "{1}: Add one mana of any color. Activate only once each turn.",
			OncePerTurn: true,
			Answers:     game.AnswerValue,
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, {T}, Sacrifice this artifact: It deals 5 damage to target creature.",
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 5)),
			Effect:  effects.DealDamageToTheTarget(5),
		}},
	})
}
