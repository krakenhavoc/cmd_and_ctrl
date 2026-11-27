package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Private Research — {U} Enchantment — Aura (Urza's Destiny, S82 OLD batch 703, #5621):
//
//	"Enchant creature
//	 At the beginning of your upkeep, you may put a page counter on this Aura.
//	 When enchanted creature dies, draw a card for each page counter on this
//	 Aura."
//
// The page counters are counted as the Aura last existed (CR 608.2h) and the
// Aura's controller draws (S82OldB703PrivateResearchDraw).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e8578152-ee97-4471-bf32-77abe5ab80f1",
		Name:         "Private Research",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.S82OldB703MayPutCounterOnThisAtUpkeep("Private Research", "page"),
			effects.WhenEnchantedCreatureDies("Private Research — draw a card for each page counter on this Aura",
				effects.S82OldB703PrivateResearchDraw),
		},
	})
}
