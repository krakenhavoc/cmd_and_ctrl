package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kor Scythemaster — Creature — Kor Soldier Ally {2}{W}, 3/1 (Oath of the Gatewatch, #5545):
//
//	"This creature has first strike as long as it's attacking."
//
// Soltari Lancer's static: first strike only while the creature is an
// attacking creature, re-judged as its attacking status changes.
//
// No simplification.
func init() {
	firstStrike := effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && target.AttackingTarget != uuid.Nil
	}, "first strike")
	firstStrike.Label = "This creature has first strike as long as it's attacking."
	firstStrike.DependsOnAttackingStatus = true
	effects.Register(effects.Spec{
		OracleID:     "21083222-12b6-4892-8078-1859140b7cc3",
		Name:         "Kor Scythemaster",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{firstStrike},
	})
}
