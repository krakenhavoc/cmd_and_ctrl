package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fire-Omen Crane — Creature — Bird Spirit {3}{R}{R}, 3/3 (Global Series Jiang Yanggu & Mu Yanling, S82 OLD batch 201, #5528):
//
//	"Flying
//	 Whenever this creature attacks, it deals 1 damage to target creature an
//	 opponent controls."
//
// Flying is the printed keyword. The attack trigger names its target as it goes
// on the stack, any creature any opponent controls (Blitz Leech's clause, not
// only the defending player's), and the Crane deals the damage as the source
// (Soldier Replica's body); a target that has become illegal is skipped (CR
// 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2306386d-5913-429e-8d8b-ba780841bf6e",
		Name:            "Fire-Omen Crane",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WheneverThisAttacks("Fire-Omen Crane — 1 damage to target creature an opponent controls", effects.B33DamageChosenTargetFromSource(1)),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
				effects.ForTargets(effects.DamageToTarget(0, 1))),
		},
	})
}
