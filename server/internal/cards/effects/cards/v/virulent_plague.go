package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Virulent Plague — Enchantment {2}{B} (Dragons of Tarkir, #5551):
//
//	"Creature tokens get -2/-2."
//
// Amulet of Safekeeping's creature-token static at -2/-2, a layer 7c change
// that applies to every creature token, whoever controls it.
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "adfba95e-c73c-415d-adea-edcd04372ba4",
		Name:         "Virulent Plague",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{effects.B16Anthem(func(target *game.Card, _ *game.Game, _ *game.Card) bool {
			return target.IsToken() && target.IsCreature()
		}, -2, -2)},
	})
}
