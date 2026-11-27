package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heart of Light — Enchantment — Aura {2}{W} (Betrayers of Kamigawa, S82 OLD b594):
//
//	"Enchant creature
//	 Prevent all damage that would be dealt to and dealt by enchanted creature."
//
// One CR 615.1a prevention over both directions (PreventAllDamageDealtToAndByAttached): the
// creature neither hurts nor is hurt, combat or not, to a player, a planeswalker or
// another creature. It reads the attachment as the damage would be dealt, so it applies to
// whichever creature the Aura enchants, whoever controls it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "634d0256-849d-4ef7-b070-b6b8f3242e00",
		Name:         "Heart of Light",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Replacements: []game.ReplacementEffect{
			effects.PreventAllDamageDealtToAndByAttached("Heart of Light: prevent all damage that would be dealt to and dealt by enchanted creature."),
		},
	})
}
