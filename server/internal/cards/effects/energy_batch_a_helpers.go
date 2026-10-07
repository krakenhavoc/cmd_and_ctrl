package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// energy_batch_a_helpers.go — shared bodies of ADR 0129 PR 1's
// "get-only" energy cards (batch A).

// scryThenGetEnergy is "Scry N, then you get M {E}" (Glassblower's
// Puzzleknot, on its enters trigger and on its activated row). The
// energy comes after the scry prompt is answered, as "then" says.
func scryThenGetEnergy(scry, energy int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		controller := item.Controller
		return Scry{Player: controller, N: scry, Then: func(g *game.Game) error {
			return g.AddPlayerCounterByForEffect(controller, controller, game.CounterEnergy, energy)
		}}.Apply(NewContext(g, item))
	}
}

// gainLifeAndGetEnergy is "you gain N life and get M {E}" (Woodweaver's
// Puzzleknot, Reservoir Walker).
func gainLifeAndGetEnergy(life, energy int) Effect {
	return Do(GainLife{Amount: life}, GetEnergy{N: energy})
}

// pumpEnchantedCreature is "Enchanted creature gets +P/+T" for an Aura
// that may enchant a noncreature (Aether Meltdown's "Enchant creature or
// Vehicle"): the change applies only while the host is a creature, which
// is what "enchanted creature" names.
func pumpEnchantedCreature(power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			if !slices.Contains(c.Types, "Creature") {
				return
			}
			c.Power += power
			c.Toughness += toughness
		},
	}
}
