package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vernal Equinox — Enchantment {3}{G} (Mercadian Masques, #5619):
//
//	"Any player may cast creature and enchantment spells as though they had
//	 flash."
//
// Yeva, Nature's Herald's flash timing for every player (Affects each
// player), over a creature OR enchantment spell (a card with at least one of
// the two types, CR 205.2a).
//
// No simplification.
func init() {
	label := "Any player may cast creature and enchantment spells as though they had flash."
	rule := effects.CastKindAsThoughFlash(game.PermissionFilter{CardTypesAny: "creature enchantment"}, label)
	rule.Affects = game.TimingAffectsEachPlayer
	effects.Register(effects.Spec{
		OracleID:     "2241ef7d-3e9d-44ee-88ac-e8ad6fb2444e",
		Name:         "Vernal Equinox",
		Completeness: effects.CompletenessFull,
		CastTimings:  []game.CastTimingRule{rule},
	})
}
