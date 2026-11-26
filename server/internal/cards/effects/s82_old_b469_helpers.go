package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b469_helpers.go — Scars of Mirrodin, S82 older sets batch 469
// (#5577): the metalcraft gates the batch's cards share. Append-only.

// OldB469Metalcraft is metalcraft read off a controller: three or more
// artifacts they control right now (CR 207.2c names it an ability word, so
// it has no rules meaning of its own; the condition is the whole ability).
func OldB469Metalcraft(g *game.Game, controller uuid.UUID) bool {
	return YouControlThreeOrMoreArtifacts(g, controller, uuid.Nil)
}

// OldB469WhenMetalcraft is the trigger condition "if you control three or
// more artifacts" for the source's controller. As the condition of a
// trigger it is the first half of an intervening "if" (CR 603.4); the
// effect wrapped by OldB469IfMetalcraft is the second.
func OldB469WhenMetalcraft(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return OldB469Metalcraft(g, source.Controller)
}

// OldB469IfMetalcraft wraps an Effect so it does nothing unless the item's
// controller has metalcraft as it resolves (CR 603.4's recheck).
func OldB469IfMetalcraft(effect Effect) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !OldB469Metalcraft(g, item.Controller) {
			return nil
		}
		return effect(g, item)
	}
}

// OldB469WhenThisEntersIfMetalcraft is "Metalcraft — When this creature
// enters, if you control three or more artifacts, …". Set Targets on the
// returned ability for a targeted body.
func OldB469WhenThisEntersIfMetalcraft(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventETB, AllOf(Self, OldB469WhenMetalcraft), label, OldB469IfMetalcraft(effect))
}

// OldB469MetalcraftActivation is "Activate only if you control three or
// more artifacts", checked as the ability is activated (CR 602.5b).
func OldB469MetalcraftActivation() ActivationCondition {
	return func(g *game.Game, controller, _ uuid.UUID) bool {
		return OldB469Metalcraft(g, controller)
	}
}

// OldB469SelfBecomesGolem is Rusted Relic's "Metalcraft — This artifact is
// a 5/5 Golem artifact creature as long as you control three or more
// artifacts": the Creature type and the Golem subtype in layer 4 and the
// base power and toughness in layer 7b, all on the one gate, which is not
// locked in (CR 611.3a). It stays an artifact.
func OldB469SelfBecomesGolem(power, toughness int) []game.StaticAbility {
	gate := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return SelfOnly(target, g, source) && somB1Metalcraft(g, source)
	}
	return []game.StaticAbility{
		{
			Label:     "Metalcraft — This artifact is a Golem artifact creature as long as you control three or more artifacts.",
			Layer:     game.Layer4Type,
			AppliesTo: gate,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if !TypeListHasEffects(c.Types, "Creature") {
					c.Types = append(c.Types, "Creature")
				}
				if !TypeListHasEffects(c.Subtypes, "Golem") {
					c.Subtypes = append(c.Subtypes, "Golem")
				}
			},
		},
		{
			Label:     "Metalcraft — This artifact is a 5/5 creature as long as you control three or more artifacts.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7B_Set,
			AppliesTo: gate,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power = power
				c.Toughness = toughness
			},
		},
	}
}
