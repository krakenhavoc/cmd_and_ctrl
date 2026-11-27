package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b570_helpers.go — shared shapes of the Guildpact Magemarks (S82
// OLD batch 570, #5599).

// OldB570YourEnchantedCreature is "a creature you control that's enchanted":
// a creature the Magemark's controller controls with an Aura attached,
// whoever controls that Aura. It is re-read on every layer pass, so the
// effect follows the Aura on and off.
func OldB570YourEnchantedCreature(target *game.Card, g *game.Game, source *game.Card) bool {
	return target.IsCreature() && target.Controller == source.Controller && IsEnchantedByAura(g, target.InstanceID)
}

// OldB570YourEnchantedCreaturePredicate is OldB570YourEnchantedCreature as a
// CardPredicate over the rule source's controller, for a block rule's scope.
func OldB570YourEnchantedCreaturePredicate(g *game.Game, caster uuid.UUID, c game.Card) bool {
	return c.IsCreature() && c.Controller == caster && IsEnchantedByAura(g, c.InstanceID)
}

// OldB570MagemarkPump is the +1/+1 every Magemark of the cycle that pumps
// gives "creatures you control that are enchanted": a layer 7c modify.
func OldB570MagemarkPump() game.StaticAbility {
	return game.StaticAbility{
		Label:     "Creatures you control that are enchanted get +1/+1.",
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: OldB570YourEnchantedCreature,
		Apply:     PumpBy(1, 1),
	}
}
