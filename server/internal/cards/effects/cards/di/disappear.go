package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Disappear — {2}{U}{U} Enchantment — Aura (Urza's Destiny, S82 OLD batch 703, #5621):
//
//	"Enchant creature
//	 {U}: Return enchanted creature and this Aura to their owners' hands."
//
// An activated ability of the Aura; the host is read first, then both cards
// are returned (S82OldB703DisappearBounce).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "039e2e39-8f7f-43f2-8448-de35ef7b8dc1",
		Name:         "Disappear",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: Return enchanted creature and this Aura to their owners' hands.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.ManaCost("{U}"),
			Effect:  effects.S82OldB703DisappearBounce,
		}},
	})
}
