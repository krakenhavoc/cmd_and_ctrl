package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const samiteBlessingGrant = "samite-blessing/prevent"

// Samite Blessing — Enchantment — Aura {W}:
//
//	"Enchant creature
//	 Enchanted creature has "{T}: The next time a source of your choice would deal damage to target creature this turn, prevent that damage.""
//
// ADR 0107 §6 (#1860): the granted row (ADR 0093) is the enchanted
// creature's own ability, so it taps, it is activated by that creature's
// controller, and "you" is that player. The shield is against the next
// instance of damage to the target from a source chosen as the ability
// resolves (CR 615.8, 609.7a).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a524ae69-2594-467d-af0a-27b9984d4300",
		Name:         "Samite Blessing",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Grants: []effects.AbilityGrant{{
			Key: samiteBlessingGrant,
			Activated: []effects.ActivatedAbility{effects.NextDamageShieldRow(
				"{T}: The next time a source of your choice would deal damage to target creature this turn, prevent that damage.",
				effects.TapCost(), effects.TargetCreature("target creature"), effects.PreventNextDamageFromChosenSource(effects.ShieldTheTarget))},
			Text: "{T}: The next time a source of your choice would deal damage to target creature this turn, prevent that damage.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(samiteBlessingGrant)},
	})
}
