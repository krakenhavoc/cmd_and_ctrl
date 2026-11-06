package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c6cd520c-106b-4ee5-bc02-6c5fbdbc78dd",
		Name:         "Barrels of Blasting Jelly",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:        ManaAbilityCost{Mana: "{1}"},
			Produced:    "{W|U|B|R|G}",
			Label:       "{1}: Add one mana of any color. Activate only once each turn.",
			OncePerTurn: true,
			Answers:     game.AnswerValue,
		}},
		Activated: []ActivatedAbility{{
			Label:   "{5}, {T}, Sacrifice this artifact: It deals 5 damage to target creature.",
			Cost:    Plus(ManaCost("{5}"), TapCost(), SacrificeThis()),
			Targets: TargetCreature("target creature"),
			Purpose: ForTargets(DamageToTarget(0, 5)),
			Effect:  DealDamageToTheTarget(5),
		}},
	})
}
