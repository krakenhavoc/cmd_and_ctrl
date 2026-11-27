package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kor Duelist — Creature — Kor Soldier {W} 1/1 (Zendikar, S82 OLD batch 492,
// #5583):
//
//	"As long as this creature is equipped, it has double strike."
//
// A layer 6 keyword grant to the creature itself while any Equipment is
// attached to it (Kitesail Apprentice's shape).
//
// No simplification.
func init() {
	whileEquipped := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.SelfOnly(target, g, source) && effects.IsEquippedCreature(g, *target)
	}
	double := effects.KeywordGrant(whileEquipped, "double strike")
	double.Label = "As long as this creature is equipped, it has double strike."
	effects.Register(effects.Spec{
		OracleID:     "758dd2fd-d3ec-4782-9a60-12b0d79719da",
		Name:         "Kor Duelist",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{double},
	})
}
