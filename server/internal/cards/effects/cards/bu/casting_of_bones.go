package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Casting of Bones — Enchantment — Aura {2}{B} (Alliances, #5634):
//
//	"Enchant creature
//	 When enchanted creature dies, draw three cards, then discard one of them."
//
// The dies trigger belongs to the Aura, so its controller draws, whoever
// controlled the creature (CR 109.5). "One of them" is one of the three cards
// just drawn, not a card that was already in hand, so the discard is a pick over
// the drawn cards (OldB816DrawThreeThenDiscardOneOfThem).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a747256-4215-4334-98ab-0c2e4ed92e47",
		Name:         "Casting of Bones",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 3},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Casting of Bones — draw three cards, then discard one of them", effects.OldB816DrawThreeThenDiscardOneOfThem),
		},
	})
}
