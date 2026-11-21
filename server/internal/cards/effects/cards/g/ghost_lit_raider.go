package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "3d772c44-a756-406d-a83f-f671a81513b0",
		Name:         "Ghost-Lit Raider",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}{R}, {T}: This creature deals 2 damage to target creature.",
				Purpose: game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 2)), Answers: game.AnswerRemove},
				Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(2),
			},
			{
				Label:   "Channel — {3}{R}, Discard this card: It deals 4 damage to target creature.",
				Purpose: game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 4)), Answers: game.AnswerRemove},
				Cost:    effects.Plus(effects.ManaCost("{3}{R}"), effects.DiscardThis()),
				Zones:   []game.ZoneKind{game.ZoneHand},
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(4),
			},
		},
	})
}
