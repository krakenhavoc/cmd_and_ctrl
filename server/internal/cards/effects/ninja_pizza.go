package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "58c615f3-8c3b-47bf-8032-809fc4232ccb",
		Name:         "Ninja Pizza",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:  grant,
			Text: "{T}, Sacrifice this artifact: Add one mana of any color.",
			Mana: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true, Sacrifice: true},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color",
			}},
		}},
		Static: []game.StaticAbility{
			GrantAbilities(foodsYouControl, grant),
		},
		Triggered: []game.TriggeredAbility{
			AtYourSecondMainPhase("Ninja Pizza — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1})),
		},
	})
}
