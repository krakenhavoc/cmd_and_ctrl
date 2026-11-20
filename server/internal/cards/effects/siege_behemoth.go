package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Siege Behemoth — Creature — Beast {5}{G}{G}, 7/4 (Commander 2014, #5553):
//
//	"Hexproof
//	 As long as this creature is attacking, for each creature you control,
//	 you may have that creature assign its combat damage as though it
//	 weren't blocked."
//
// Pride of Lions' flag (Characteristic.AssignsAsThoughUnblocked, CR
// 510.1c) on every creature you control while the Behemoth is attacking, a
// layer 6 static that reads the attacking status (Blood Petal Celebrant's
// DependsOnAttackingStatus). The "may" is the engine's per-creature choice
// as each blocked creature assigns its damage. Once the Behemoth is removed
// from combat or leaves, the others assign as usual.
//
// No simplification.
func init() {
	assigns := game.StaticAbility{
		Layer: game.Layer6Ability,
		Label: "As long as this creature is attacking, for each creature you control, you may have that creature assign its combat damage as though it weren't blocked.",
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return source.AttackingTarget != uuid.Nil && target.IsCreature() && target.Controller == source.Controller
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.AssignsAsThoughUnblocked = true
		},
		DependsOnAttackingStatus: true,
	}
	Register(Spec{
		OracleID:        "cadf7be1-4a74-4163-bb2f-6bfbbb557973",
		Name:            "Siege Behemoth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		Static:          []game.StaticAbility{assigns},
	})
}
