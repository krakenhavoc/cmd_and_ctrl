package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cloak of Invisibility — Enchantment — Aura {U} (Mirage, S82 OLD batch 811,
// #5633):
//
//	"Enchant creature
//	 Enchanted creature has phasing and can't be blocked except by Walls. (It
//	 phases in or out before its controller untaps during each of their untap
//	 steps. While it's phased out, it's treated as though it doesn't exist.)"
//
// Invisibility's block rule, scoped to the enchanted creature (OnAttached), and
// a layer 6 grant of phasing (CR 702.26a), which the engine honours on its own
// once the creature has it. Both end the moment the Aura does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68b06917-7f83-47d4-a466-4e6698cae56d",
		Name:         "Cloak of Invisibility",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached(game.KeywordPhasing)},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedExceptBy(effects.OnAttached(), effects.OfCreatureType("Wall"), "Walls"),
		},
	})
}
