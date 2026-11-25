package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spiteful Motives — Enchantment — Aura {3}{R} (Shadows over Innistrad, S82
// OLD):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets +3/+0 and has first strike."
//
// Minion's Return's flash on an Aura, a flat pump and a granted keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a8f259b5-ef01-4f30-baa1-1b2a5419e61e",
		Name:            "Spiteful Motives",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 0),
			effects.GrantToAttached("first strike"),
		},
	})
}
