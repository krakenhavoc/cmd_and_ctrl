package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a524ae69-2594-467d-af0a-27b9984d4300",
		Name:         "Samite Blessing",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Grants: []AbilityGrant{{
			Key: samiteBlessingGrant,
			Activated: []ActivatedAbility{nextDamageShieldRow(
				"{T}: The next time a source of your choice would deal damage to target creature this turn, prevent that damage.",
				TapCost(), TargetCreature("target creature"), PreventNextDamageFromChosenSource(ShieldTheTarget))},
			Text: "{T}: The next time a source of your choice would deal damage to target creature this turn, prevent that damage.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToAttached(samiteBlessingGrant)},
	})
}
