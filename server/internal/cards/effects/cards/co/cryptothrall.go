package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cryptothrall — Artifact Creature — Construct {4}, 3/3 (Warhammer 40,000
// Commander, S82 GAP batch 187, #5948):
//
//	"Protector — Other artifact creatures you control have hexproof."
//
// A layer 6 grant over the controller's other artifact creatures, read live
// on every pass (CR 611.3a), Shalai's shape.
//
// No simplification.
func init() {
	grant := effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID != source.InstanceID && target.Controller == source.Controller &&
			target.IsCreature() && target.IsArtifact()
	}, effects.KeywordHexproof)
	grant.Label = "Other artifact creatures you control have hexproof."
	effects.Register(effects.Spec{
		OracleID:     "e26f0fcd-d6d2-4e7b-82dd-711bc0b9dad3",
		Name:         "Cryptothrall",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{grant},
	})
}
