package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unquestioned Authority — Enchantment — Aura {2}{W} (Jumpstart 2022, #5948):
//
//	"Enchant creature
//	 When this Aura enters, draw a card.
//	 Enchanted creature has protection from creatures."
//
// Spirit Mantle's protection grant (CR 702.16a: no creature can block it,
// creature damage is prevented, creatures' abilities can't target it) and a
// draw as the Aura enters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d50266f-5a2f-4a48-ab15-87e733624fd4",
		Name:         "Unquestioned Authority",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 1},
		Static:       []game.StaticAbility{effects.GrantToAttached("protection from creatures")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Unquestioned Authority — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
