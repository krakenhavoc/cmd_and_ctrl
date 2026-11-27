package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Earth Surge — Enchantment {3}{G} (Guildpact):
//
//	"Each land gets +2/+2 as long as it's a creature."
//
// Depala's "gets +1/+1 as long as it's a creature": a layer 7c bonus for every
// land, whoever controls it, that is a creature right now, read after layer 4
// has settled what is a creature, so a land that stops being one loses it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb8fae53-3b06-41d8-baa0-638820a4b550",
		Name:         "Earth Surge",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Each land gets +2/+2 as long as it's a creature.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.IsLand() && target.IsCreature()
			},
			Apply: effects.PumpBy(2, 2),
		}},
	})
}
