package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ninja Pizza — Enchantment {2}{G} (Teenage Mutant Ninja Turtles
// Commander, #5011):
//
//	"Foods you control have '{T}, Sacrifice this artifact: Add one mana
//	 of any color.'
//	 At the beginning of your second main phase, create a Food token."
//
// The first line is Night of the Sweets' Revenge's layer-6 grant to
// "Foods you control" (foodsYouControl, any Food artifact, token or not)
// with Goldspan Dragon's mana ability: tap and sacrifice the Food itself
// for one mana of any color. A Food sacrificed this way cannot also pay
// its own life-gain ability (the card's ruling): each is a separate
// activation with its own sacrifice. The second line makes the CR 111.10b
// Food at the turn's second main phase, not an added third
// (AtYourSecondMainPhase).
//
// No simplification.
func init() {
	const grant = "ninja-pizza/food-any-color"
	effects.Register(effects.Spec{
		OracleID:     "58c615f3-8c3b-47bf-8032-809fc4232ccb",
		Name:         "Ninja Pizza",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:  grant,
			Text: "{T}, Sacrifice this artifact: Add one mana of any color.",
			Mana: []effects.ManaAbility{{
				Cost:     effects.ManaAbilityCost{Tap: true, Sacrifice: true},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color",
			}},
		}},
		Static: []game.StaticAbility{
			effects.GrantAbilities(effects.FoodsYouControl, grant),
		},
		Triggered: []game.TriggeredAbility{
			effects.AtYourSecondMainPhase("Ninja Pizza — create a Food token", effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
		},
	})
}
