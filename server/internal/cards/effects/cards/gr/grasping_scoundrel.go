package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grasping Scoundrel — Creature — Human Pirate {B}, 1/1 (Rivals of
// Ixalan, #5531):
//
//	"This creature gets +1/+0 as long as it's attacking."
//
// Kitesail Corsair's "as long as it's attacking" gate on a layer 7c pump:
// the static depends on attacking status, so it is re-read the moment the
// creature is declared an attacker and ends when it is removed from
// combat.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d8c3298f-da96-46ee-8b46-ce1d0df95f62",
		Name:         "Grasping Scoundrel",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:                    "This creature gets +1/+0 as long as it's attacking.",
			Layer:                    game.Layer7PT,
			SubLayer:                 game.SubLayer7C_Modify,
			DependsOnAttackingStatus: true,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && target.AttackingTarget != uuid.Nil
			},
			Apply: effects.PumpBy(1, 0),
		}},
	})
}
