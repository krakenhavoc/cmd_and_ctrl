package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const auraFluxGrant = "aura-flux/upkeep-tax"

// Aura Flux — Enchantment {2}{U}:
//
//	"Other enchantments have "At the beginning of your upkeep, sacrifice
//	 this enchantment unless you pay {2}.""
//
// A layer-6 grant (CR 613.1f) of one triggered ability to every other
// enchantment, whoever controls it: each one's controller pays at their own
// upkeep, or sacrifices that enchantment. A recipient that stops being an
// enchantment, or Aura Flux leaving, ends the grant.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "128304af-4555-4276-bf26-638017beb3ff",
		Name:         "Aura Flux",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: auraFluxGrant,
			Triggered: []game.TriggeredAbility{
				effects.AtYourUpkeep("Aura Flux — sacrifice this enchantment unless you pay {2}", func(g *game.Game, item *game.StackItem) error {
					return effects.UpkeepPayUnless{
						Chooser:  item.Controller,
						Cost:     "{2}",
						Question: "Aura Flux — pay {2} or sacrifice this enchantment?",
						Else:     effects.UpkeepSacrificeSource,
					}.Apply(effects.NewContext(g, item))
				}),
			},
			Text: "At the beginning of your upkeep, sacrifice this enchantment unless you pay {2}.",
		}},
		Static: []game.StaticAbility{
			effects.GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.HasCardType("enchantment") && target.InstanceID != source.InstanceID
			}, auraFluxGrant),
		},
	})
}
