package cardsv

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vivid Flying Fish — Creature — Fish Lizard {1}{U}, 1/1 (Global Series Jiang Yanggu & Mu Yanling, #5528):
//
//	"This creature has flying as long as it's attacking."
//
// Blood Petal Celebrant's attacking-status grant, for flying.
//
// No simplification.
func init() {
	flying := effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && target.AttackingTarget != uuid.Nil
	}, "flying")
	flying.Label = "This creature has flying as long as it's attacking."
	flying.DependsOnAttackingStatus = true
	effects.Register(effects.Spec{
		OracleID:     "e8ec9ba4-dc6b-437d-863f-5ec37f4ae9cd",
		Name:         "Vivid Flying Fish",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{flying},
	})
}
