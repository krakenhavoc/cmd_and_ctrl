package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunspear Shikari — Creature — Cat Soldier {1}{W}, 2/2 (Foundations Jumpstart):
//
//	"As long as this creature is equipped, it has first strike and lifelink."
//
// Armed Assailant's "as long as this creature is equipped" gate
// (isEquippedCreature, read on every layer recompute) on two layer 6
// keyword grants to the creature alone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c49db060-2d4b-4dc8-8e54-7606e079c9fe",
		Name:         "Sunspear Shikari",
		Completeness: effects.CompletenessFull,
		Static:       sunspearShikariStatics(),
	})
}

func sunspearShikariStatics() []game.StaticAbility {
	whileEquipped := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.SelfOnly(target, g, source) && effects.IsEquippedCreature(g, *target)
	}
	fs := effects.KeywordGrant(whileEquipped, "first strike")
	fs.Label = "As long as this creature is equipped, it has first strike."
	ll := effects.KeywordGrant(whileEquipped, "lifelink")
	ll.Label = "As long as this creature is equipped, it has lifelink."
	return []game.StaticAbility{fs, ll}
}
