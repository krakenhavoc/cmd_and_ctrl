package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grisly Transformation — {2}{B} Enchantment — Aura (Born of the Gods):
//
//	"Enchant creature
//	When this Aura enters, draw a card.
//	Enchanted creature has intimidate."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0bb359bb-a1db-4535-8245-097ee989f4f5",
		Name:         "Grisly Transformation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 1},
		Static:       []game.StaticAbility{effects.GrantToAttached("intimidate")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Grisly Transformation — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
