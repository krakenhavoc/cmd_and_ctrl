package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Raph's Bravado — {1}{R} Enchantment (Teenage Mutant Ninja Turtles
// Commander, #5011):
//
//	"During your turn, attacking creatures get +1/+0."
//
// A layer-7c static over every attacking creature, whoever controls it,
// that holds only while its controller is the active player. It is re-read
// on every pass (DependsOnAttackingStatus), so a creature gets the bonus
// as it is declared an attacker and loses it as combat ends or it is
// removed from combat (CR 506.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7991aff6-0b85-41a3-bc31-09811d6fa353",
		Name:         "Raph's Bravado",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:                    "During your turn, attacking creatures get +1/+0.",
			Layer:                    game.Layer7PT,
			SubLayer:                 game.SubLayer7C_Modify,
			DependsOnAttackingStatus: true,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.AttackingTarget != uuid.Nil && isActivePlayer(g, source.Controller)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
			},
		}},
	})
}
