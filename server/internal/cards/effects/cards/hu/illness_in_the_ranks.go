package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Illness in the Ranks — Enchantment {B} (Gatecrash, S82 OLD #5565):
//
//	"Creature tokens get -1/-1."
//
// Virulent Plague's static at -1/-1: a layer 7c change to every creature token,
// whoever controls it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "51abc95c-bd92-46f7-a0a1-6e9f8e0dda87",
		Name:         "Illness in the Ranks",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{effects.B16Anthem(func(target *game.Card, _ *game.Game, _ *game.Card) bool {
			return target.IsToken() && target.IsCreature()
		}, -1, -1)},
	})
}
