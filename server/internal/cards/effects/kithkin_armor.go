package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kithkin Armor — Enchantment — Aura {W}:
//
//	"Enchant creature
//	 Enchanted creature can't be blocked by creatures with power 3 or greater.
//	 Sacrifice this Aura: The next time a source of your choice would deal damage to enchanted creature this turn, prevent that damage."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a source chosen as the ability resolves (CR 615.8, 609.7a).
// The Aura is sacrificed to pay for it, so "enchanted creature" is the
// creature it was attached to as it last existed (CR 608.2h); a creature
// that has left the battlefield by resolution gets no shield.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "35e2730e-b3aa-4080-aae8-7b3ad1e59848",
		Name:         "Kithkin Armor",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		BlockRules:   []game.BlockRule{CantBeBlockedBy(OnAttached(), PowerGE(3), "creatures with power 3 or greater")},
		Activated: []ActivatedAbility{nextDamageShieldRow(
			"Sacrifice this Aura: The next time a source of your choice would deal damage to enchanted creature this turn, prevent that damage.",
			SacrificeThis(), nil, PreventNextDamageFromChosenSource(ShieldEnchantedCreature))},
	})
}
