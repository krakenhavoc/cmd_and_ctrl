package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Adanto Vanguard — Creature — Vampire Soldier {1}{W}, 1/1 (Ixalan, S82
// older sets):
//
//	"As long as this creature is attacking, it gets +2/+0.
//	 Pay 4 life: This creature gains indestructible until end of turn."
//
// The static is Elturel Survivors' attacking-status static (layer 7c, it
// depends on the attacking status); the ability's cost is life (CR 119.4: a
// player can pay only the life they have).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91c380f6-2c41-48b7-9f07-104e8659e397",
		Name:         "Adanto Vanguard",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "As long as this creature is attacking, it gets +2/+0.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && target.AttackingTarget != uuid.Nil
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 2
			},
			DependsOnAttackingStatus: true,
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "Pay 4 life: This creature gains indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.PayLife(4),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("indestructible", "Adanto Vanguard — indestructible until end of turn"),
		}},
	})
}
