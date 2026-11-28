package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Black Scarab — Enchantment — Aura {W} (Ice Age, S82 OLD b828, #5636):
//
//	"Enchant creature
//	 Enchanted creature can't be blocked by black creatures.
//	 Enchanted creature gets +2/+2 as long as an opponent controls a black
//	 permanent."
//
// Air Bladder's block rule on the enchanted creature (CantBeBlockedBy with
// OnAttached) and OldB828PumpAttachedWhileOpponentControls for the
// conditional pump, read live on every recompute. Shared with the other
// Scarab.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2774d8b3-dc5f-4281-b4c5-cb003fef0ab2",
		Name:         "Black Scarab",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.OldB828PumpAttachedWhileOpponentControls(effects.OfColor("B"), 2, 2, "Enchanted creature gets +2/+2 as long as an opponent controls a black permanent."),
		},
		BlockRules: []game.BlockRule{effects.CantBeBlockedBy(effects.OnAttached(), effects.OfColor("B"), "black creatures")},
	})
}
