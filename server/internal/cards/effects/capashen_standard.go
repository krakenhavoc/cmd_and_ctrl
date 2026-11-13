package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Capashen Standard — Enchantment — Aura {W} (Urza's Destiny, #5621):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1.
//	 {2}, Sacrifice this Aura: Draw a card."
//
// A +1/+1 Aura that cashes itself in through the shared
// udsB2SacrificeThisDrawRow: sacrificing it is a cost, so the draw
// happens and the creature loses the bonus at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "75510429-41bb-409e-b6fe-04a8bb174c6b",
		Name:         "Capashen Standard",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{PumpAttached(1, 1)},
		Activated:    []ActivatedAbility{udsB2SacrificeThisDrawRow("Aura")},
	})
}
