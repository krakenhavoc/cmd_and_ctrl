package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// SacrificeEnchantedCreature is "Sacrifice enchanted creature" as an activation
// cost (Bloodfire Infusion, S82 OLD batch 631, #5614): the creature the Aura is
// attached to, no other. It is not chosen from a menu of the activator's
// creatures: the clause admits only the permanent the paying Aura is attached to
// (game.TargetSpec.SourceAttachedTo), so the payment validator, the legal
// enumerator and the protocol view all agree. The sacrifice is a cost, paid as
// the ability is activated (CR 602.2b), and the creature's last-known power is
// what the effect reads (CR 608.2h).
func SacrificeEnchantedCreature() game.AbilityCost {
	spec := SacrificeSpec("the enchanted creature", Creature())
	spec.SourceAttachedTo = true
	return game.AbilityCost{SacrificeOther: spec}
}

// OldB631EnchantedCantBlock is "Enchanted creature can't block this turn" for an
// Aura's activated ability (Manacles of Decay): the creature the Aura is
// attached to as the ability resolves, and nothing if it is no longer attached
// (CR 608.2b).
func OldB631EnchantedCantBlock(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		host := EnchantedCreatureOf(ctx)
		if host == uuid.Nil {
			return nil
		}
		return RestrictUntilEOT{Target: host, Restrictions: game.CantBlock, Label: label}.Apply(ctx)
	}
}
