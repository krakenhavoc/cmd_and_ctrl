package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bad Moon — Enchantment {1}{B} (Foundations Commander):
//
//	"Black creatures get +1/+1."
//
// One layer 7c static over every black creature at the table, whoever
// controls it (the card prints no controller clause). Colour is read
// through the layered view and the source's word lens (TextColor), as
// Breath of Dreams reads green.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fc5d3341-cbce-49e5-93cc-8add92479dca",
		Name:         "Bad Moon",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.HasColor(source.TextColor("B"))
			}, 1, 1),
		},
	})
}
