package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bequeathal — Enchantment — Aura {G} (Exodus, S82 OLD b739, #5626):
//
//	"Enchant creature
//	 When enchanted creature dies, you draw two cards."
//
// The dies trigger is harvested while the Aura is still attached (Infectious
// Bloodlust's read); "you" is the Aura's controller, who draws two cards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a9054ce-7276-4030-901a-cf6aa18f2b62",
		Name:         "Bequeathal",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Bequeathal — you draw two cards", effects.Do(effects.DrawCards{N: 2})),
		},
	})
}
