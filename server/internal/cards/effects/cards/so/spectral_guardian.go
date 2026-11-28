package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spectral Guardian — Creature — Spirit {2}{W}{W}, 2/3 (Mirage, S82 OLD batch
// 810, #5633):
//
//	"As long as this creature is untapped, noncreature artifacts have shroud.
//	 (They can't be the targets of spells or abilities.)"
//
// A layer 6 grant over every noncreature artifact, whoever controls it, read
// live on every pass (CR 611.3a) and applying only while the Guardian is
// untapped (Cryptothrall's grant shape).
//
// No simplification.
func init() {
	grant := effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return !source.Tapped && target.IsArtifact() && !target.IsCreature()
	}, effects.KeywordShroud)
	grant.Label = "As long as this creature is untapped, noncreature artifacts have shroud."
	effects.Register(effects.Spec{
		OracleID:     "87d734bc-4e4d-4013-91ed-c6bf1dde824e",
		Name:         "Spectral Guardian",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{grant},
	})
}
