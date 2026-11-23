package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tinder Wall — Creature — Plant Wall {G}, 0/3 (Mystery Booster 2):
//
//	"Defender
//	 Sacrifice this creature: Add {R}{R}.
//	 {R}, Sacrifice this creature: It deals 2 damage to target creature
//	 it's blocking."
//
// The first ability is a mana ability with a sacrifice cost (Blood Pet's
// shape, CR 605.3b), with no {T} in the cost, so summoning sickness never
// applies. The second targets, so it uses the stack: "target creature it's
// blocking" is BlockedBySource (Goblin Snowman's clause), judged at announce
// and again at resolution, and the Wall deals the damage as it last existed
// once it has been sacrificed to pay (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dd2f698c-b3b4-4dc8-9950-9be03ca7dc9e",
		Name:            "Tinder Wall",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Sacrifice: true},
			Produced: "{R}{R}",
			Label:    "Sacrifice this creature: Add {R}{R}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, Sacrifice this creature: It deals 2 damage to target creature it's blocking.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Cost:    game.AbilityCost{Mana: "{R}", SacrificeSelf: true},
			Targets: effects.BlockedBySource(effects.TargetCreature("target creature it's blocking")),
			Effect:  effects.S82GapB21DamageToBlocked(2),
		}},
	})
}
