package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sheltering Boughs — {2}{G} Enchantment — Aura (Innistrad: Crimson Vow,
// S82):
//
//	"Enchant creature
//	 When this Aura enters, draw a card.
//	 Enchanted creature gets +1/+3."
//
// Angelic Gift's enters-draw with PumpAttached for the layer 7c change.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b6bde4eb-15f0-4e6e-9d25-459a2f75ccf2",
		Name:         "Sheltering Boughs",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 1},
		Static:       []game.StaticAbility{PumpAttached(1, 3)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sheltering Boughs — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
