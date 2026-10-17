package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Instigator Gang // Wildblood Pack — {3}{R} Creature — Human Werewolf
// 2/3 // Creature — Werewolf 5/5 (#4071, ADR 0132 amendment 2026-10-16):
//
//	Front: "Attacking creatures you control get +1/+0.
//	        At the beginning of each upkeep, if no spells were cast last
//	        turn, transform this creature."
//	Back:  "Trample
//	        Attacking creatures you control get +3/+0.
//	        At the beginning of each upkeep, if a player cast two or more
//	        spells last turn, transform this creature."
//
// Each face's boost is a layer 7c static over your attacking creatures
// (the Gang itself too, when it attacks), Watchdog's shape: it reads
// attacking status, so it declares DependsOnAttackingStatus.
//
// No simplification.
func init() {
	const oracle = "27c2bb87-7e50-4a80-b41c-ff0237a3d648"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Instigator Gang",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{attackingCreaturesYouControlGetPower(1)},
		Triggered: []game.TriggeredAbility{
			TransformAtUpkeepIfNoSpellsLastTurn("Instigator Gang"),
		},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Wildblood Pack",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static:          []game.StaticAbility{attackingCreaturesYouControlGetPower(3)},
		Triggered: []game.TriggeredAbility{
			TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Wildblood Pack"),
		},
	})
}

// attackingCreaturesYouControlGetPower is "Attacking creatures you
// control get +N/+0".
func attackingCreaturesYouControlGetPower(n int) game.StaticAbility {
	return game.StaticAbility{
		Label:                    "Attacking creatures you control get " + signedPT(n, 0) + ".",
		Layer:                    game.Layer7PT,
		SubLayer:                 game.SubLayer7C_Modify,
		DependsOnAttackingStatus: true,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.Controller == source.Controller && target.AttackingTarget != uuid.Nil
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += n
		},
	}
}
