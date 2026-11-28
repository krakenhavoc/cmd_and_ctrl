package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cooperation — Enchantment — Aura {2}{W} (Ice Age, S82 OLD b828, #5636):
//
//	"Enchant creature
//	 Enchanted creature has banding."
//
// Banding is the engine's (game/banding.go, ADR 0224), granted the way Arcane
// Flight grants flying.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "19f7a0f2-aa3f-4253-be44-b08056918a26",
		Name:         "Cooperation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached(game.KeywordBanding)},
	})
}
