package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghost-Lit Raider — Creature — Spirit {2}{R}, 2/1 (Saviors of Kamigawa,
// #5601):
//
//	"{2}{R}, {T}: This creature deals 2 damage to target creature.
//	 Channel — {3}{R}, Discard this card: It deals 4 damage to target
//	 creature."
//
// The channel row works only from the hand, the discard part of its cost;
// "it" is the discarded card, the stack item's source, as Twinshot
// Sniper's is. A target that left in response is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3d772c44-a756-406d-a83f-f671a81513b0",
		Name:         "Ghost-Lit Raider",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{2}{R}, {T}: This creature deals 2 damage to target creature.",
				Purpose: game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 2)), Answers: game.AnswerRemove},
				Cost:    Plus(ManaCost("{2}{R}"), TapCost()),
				Targets: TargetCreature("target creature"),
				Effect:  sourceDealsDamageToEachLegalTarget(2),
			},
			{
				Label:   "Channel — {3}{R}, Discard this card: It deals 4 damage to target creature.",
				Purpose: game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 4)), Answers: game.AnswerRemove},
				Cost:    Plus(ManaCost("{3}{R}"), DiscardThis()),
				Zones:   []game.ZoneKind{game.ZoneHand},
				Targets: TargetCreature("target creature"),
				Effect:  sourceDealsDamageToEachLegalTarget(4),
			},
		},
	})
}
