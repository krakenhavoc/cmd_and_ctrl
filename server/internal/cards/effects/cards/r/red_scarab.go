package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Red Scarab — Enchantment — Aura {W} (Ice Age, S82 OLD b828, #5636):
//
//	"Enchant creature
//	 Enchanted creature can't be blocked by red creatures.
//	 Enchanted creature gets +2/+2 as long as an opponent controls a red
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
		OracleID:     "380b6a48-c364-4939-be0b-df542c63a5ff",
		Name:         "Red Scarab",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.OldB828PumpAttachedWhileOpponentControls(effects.OfColor("R"), 2, 2, "Enchanted creature gets +2/+2 as long as an opponent controls a red permanent."),
		},
		BlockRules: []game.BlockRule{effects.CantBeBlockedBy(effects.OnAttached(), effects.OfColor("R"), "red creatures")},
	})
}
