package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prismatic Omen — Enchantment {1}{G} (Enchanting Tales):
//
//	"Lands you control are every basic land type in addition to their other
//	 types."
//
// Dryad of the Ilysian Grove's layer-4 static: every land you control gains
// Plains, Island, Swamp, Mountain and Forest, so each taps for any of the five
// colours through the intrinsic mana abilities of CR 305.6.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d90882ae-6572-4214-8d9c-0bdfb75f9be2",
		Name:         "Prismatic Omen",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "Lands you control are every basic land type in addition to their other types.",
			Layer: game.Layer4Type,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsLand() && target.Controller == source.Controller
			},
			Apply: effects.DryadOfTheIlysianGroveGrantAllBasicTypes,
		}},
	})
}
