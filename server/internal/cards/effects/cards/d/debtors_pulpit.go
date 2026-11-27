package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Debtor's Pulpit — Enchantment — Aura {4}{W} (Gatecrash, S82 OLD b407,
// #5565):
//
//	"Enchant land
//	 Enchanted land has "{T}: Tap target creature.""
//
// Underworld Connections' granted row with a target: the land's controller
// activates it, and the creature is tapped if it is still a legal target as the
// ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	const grant = "debtors-pulpit/tap"
	effects.Register(effects.Spec{
		OracleID:     "541c6564-ec39-4b14-8c02-dd6a5d122ed3",
		Name:         "Debtor's Pulpit",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Tap target creature.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.TapEachLegalTargetBody,
			}},
			Text: "{T}: Tap target creature.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(grant)},
	})
}
