package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pattern of Rebirth — {3}{G} Enchantment — Aura (Urza's Destiny, S82 OLD batch 703, #5621):
//
//	"Enchant creature
//	 When enchanted creature dies, that creature's controller may search their
//	 library for a creature card, put that card onto the battlefield, then
//	 shuffle."
//
// The searcher is the creature's last-known controller (CR 603.10a); the
// search is optional (S82OldB703PatternOfRebirth).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b8856f8-3cd1-450c-a71b-eb0026e8e179",
		Name:         "Pattern of Rebirth",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Pattern of Rebirth — that creature's controller may search for a creature card and put it onto the battlefield",
				effects.S82OldB703PatternOfRebirth),
		},
	})
}
