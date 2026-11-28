package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spirit of the Night — Legendary Creature — Demon Spirit {6}{B}{B}{B}, 6/5
// (Mirage, S82 OLD b806):
//
//	"Flying, trample, haste, protection from black
//	 Spirit of the Night has first strike as long as it's attacking."
//
// The printed keywords are engine data; the first strike is Kor Scythemaster's
// attacking-status static, re-read the moment it is declared an attacker.
//
// No simplification.
func init() {
	firstStrike := effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && target.AttackingTarget != uuid.Nil
	}, "first strike")
	firstStrike.Label = "Spirit of the Night has first strike as long as it's attacking."
	firstStrike.DependsOnAttackingStatus = true
	effects.Register(effects.Spec{
		OracleID:        "73016220-2436-4072-96dd-65c0cb05fff4",
		Name:            "Spirit of the Night",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample", "haste", game.ProtectionFromColor("B")},
		Static:          []game.StaticAbility{firstStrike},
	})
}
