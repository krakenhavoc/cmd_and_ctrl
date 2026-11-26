package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Security Blockade — Enchantment — Aura {2}{W} (Return to Ravnica, S82 OLD):
//
//	"Enchant land
//	 When this Aura enters, create a 2/2 white Knight creature token with
//	 vigilance.
//	 Enchanted land has "{T}: Prevent the next 1 damage that would be dealt to
//	 you this turn.""
//
// The enters trigger makes the Knight for the Aura's controller. The land's
// controller, who may be an opponent, taps it for a one-damage shield on
// themselves (CR 615.7), which ends at cleanup.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b4b397a9-85cb-4bdf-8c2e-a151a02a4208",
		Name:         "Security Blockade",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Security Blockade — create a 2/2 white Knight creature token with vigilance",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 white Knight with vigilance"), N: 1})),
		},
		Grants: []effects.AbilityGrant{{
			Key: securityBlockadeGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Prevent the next 1 damage that would be dealt to you this turn.",
				Purpose: game.Purpose{Answers: game.AnswerPrevent},
				Cost:    effects.TapCost(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.PreventNextDamage{Target: item.Controller, Amount: 1, Label: "Security Blockade: prevent the next 1 damage"}.Apply(effects.NewContext(g, item))
				},
			}},
			Text: "{T}: Prevent the next 1 damage that would be dealt to you this turn.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(securityBlockadeGrant)},
	})
}

// securityBlockadeGrant is the granted ability's bundle key: an on-disk
// identity, never renamed, never reused.
const securityBlockadeGrant = "security-blockade/prevent-one"
