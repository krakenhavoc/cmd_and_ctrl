package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "dd2f698c-b3b4-4dc8-9950-9be03ca7dc9e",
		Name:            "Tinder Wall",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Sacrifice: true},
			Produced: "{R}{R}",
			Label:    "Sacrifice this creature: Add {R}{R}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{R}, Sacrifice this creature: It deals 2 damage to target creature it's blocking.",
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Cost:    game.AbilityCost{Mana: "{R}", SacrificeSelf: true},
			Targets: BlockedBySource(TargetCreature("target creature it's blocking")),
			Effect:  s82GapB21DamageToBlocked(2),
		}},
	})
}
