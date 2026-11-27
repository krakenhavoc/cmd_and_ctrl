package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wreath of Geists — Enchantment — Aura {G}:
//
//	"Enchant creature
//	 Enchanted creature gets +X/+X, where X is the number of creature cards in
//	 your graveyard."
//
// X is read on every layer pass, so it follows the graveyard; "your" is the
// Aura's controller's, not the host's (CR 109.5).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b11c31ac-9a52-4faa-b03e-5a78129fd43f",
		Name:         "Wreath of Geists",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttachedPer(1, 1, effects.OldB444CreatureCardsInYourGraveyard)},
	})
}
