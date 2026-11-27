package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Treetop Bracers — Enchantment — Aura {1}{G} (Nemesis, S82 OLD batch 680,
// #5618):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and can't be blocked except by creatures
//	 with flying."
//
// The pump is a layer 7c modify of the enchanted creature; the evasion is a
// block rule on it that refuses every blocker without flying, read as blocks
// are declared (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ba71eb8a-f9f8-4608-88d8-963a1770d8d4",
		Name:         "Treetop Bracers",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedExceptBy(effects.OnAttached(), effects.HasKeyword("flying"), "creatures with flying"),
		},
	})
}
