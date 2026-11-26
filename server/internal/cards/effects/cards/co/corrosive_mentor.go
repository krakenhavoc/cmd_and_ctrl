package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Corrosive Mentor — Creature — Elemental Rogue {2}{B}, 1/3 (Shadowmoor, S82
// OLD b530, #5591):
//
//	"Black creatures you control have wither. (They deal damage to creatures in
//	 the form of -1/-1 counters.)"
//
// Massacre Girl's layer-6 keyword grant narrowed to black creatures, the Mentor
// included, read on post-layer colour.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "380a1fbf-68bc-4456-9a67-f01c0174944c",
		Name:         "Corrosive Mentor",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.CreaturesYouControl(target, g, source) && target.HasColorWord("B")
			}, game.KeywordWither),
		},
	})
}
