package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Security Bypass — Enchantment — Aura {1}{U} (Streets of New Capenna,
// #5491):
//
//	"Enchant creature
//	 As long as enchanted creature is attacking alone, it can't be blocked.
//	 Enchanted creature has "Whenever this creature deals combat damage to a
//	 player, it connives." (Its controller draws a card, then discards a
//	 card. If they discarded a nonland card, they put a +1/+1 counter on
//	 this creature.)"
//
// The evasion is Gutter Shortcut's block rule (CR 506.5 attacking alone,
// read when blockers are declared) on the enchanted creature. The quoted
// trigger is a granted bundle (ADR 0093) so the enchanted creature, not the
// Aura, connives, and its controller draws and discards.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cb660c6d-dc8b-4779-8613-6acf19418baa",
		Name:         "Security Bypass",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		BlockRules:   []game.BlockRule{CantBeBlockedWhile(OnAttached(), AttackingAlone())},
		Grants: []AbilityGrant{{
			Key: securityBypassConnive,
			Triggered: []game.TriggeredAbility{
				On(game.EventDealDamage, ThisDealtCombatDamageToAPlayer,
					"Security Bypass — it connives", Do(Instruction(sourceConnives))),
			},
			Text: "Whenever this creature deals combat damage to a player, it connives.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToAttached(securityBypassConnive)},
	})
}

const securityBypassConnive = "security-bypass/connive"
