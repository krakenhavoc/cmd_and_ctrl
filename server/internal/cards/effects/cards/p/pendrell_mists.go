package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pendrell Mists — Enchantment {3}{U} (Weatherlight, S82 older sets batch 778, #5629):
//
//	"All creatures have "At the beginning of your upkeep, sacrifice this creature unless you
//	 pay {1}.""
//
// Plague Sliver's granted trigger over every creature on the battlefield (ADR 0093), not a tribe:
// each creature carries its own upkeep trigger, controlled by its own controller, so each pays {1}
// or sacrifices that creature at its controller's upkeep (Elder Dragon tax, UpkeepSacrificeUnlessYouPay).
//
// No simplification.
func init() {
	const key = "pendrell-mists/upkeep"
	effects.Register(effects.Spec{
		OracleID:     "b18b8415-a0b1-459d-8c6b-896381184935",
		Name:         "Pendrell Mists",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       key,
			Triggered: []game.TriggeredAbility{effects.UpkeepSacrificeUnlessYouPay("this creature", "{1}")},
			Text:      "At the beginning of your upkeep, sacrifice this creature unless you pay {1}.",
		}},
		Static: []game.StaticAbility{
			effects.GrantAbilities(func(target *game.Card, _ *game.Game, _ *game.Card) bool { return target.IsCreature() }, key),
		},
	})
}
