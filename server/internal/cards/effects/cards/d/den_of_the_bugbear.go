package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Den of the Bugbear — Land (Adventures in the Forgotten Realms):
//
//	"If you control two or more other lands, this land enters tapped.
//	 {T}: Add {R}.
//	 {3}{R}: Until end of turn, this land becomes a 3/2 red Goblin creature
//	 with "Whenever this creature attacks, create a 1/1 red Goblin creature
//	 token that's tapped and attacking." It's still a land."
//
// The animation is afrB17Manland (one effect, layers 4, 5, 6 and 7b); the
// quoted trigger is an ability bundle granted with it (ADR 0093), so it
// belongs to the land and goes away with the effect. The token is created
// tapped and attacking the player the land attacked (CR 508.4: it was never
// declared as an attacker, so nothing triggers off its attack). The land's
// own {T} ability can't be used while it is a creature that came under its
// controller's control this turn (CR 302.6), which the engine supplies.
//
// No simplification.
const denOfTheBugbearAttackToken = "den-of-the-bugbear/attack-token"

func init() {
	effects.Register(effects.Spec{
		OracleID:      "f451b8f0-1ff5-4e8d-9f30-9352d83ed687",
		Name:          "Den of the Bugbear",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.AfrB17EntersTappedWithTwoOtherLands()},
		ManaAbilities: effects.AfrB17TapForMana("R"),
		Grants: []effects.AbilityGrant{{
			Key: denOfTheBugbearAttackToken,
			Triggered: []game.TriggeredAbility{
				effects.WhenThisAttacksTokensTappedAndAttacking("Den of the Bugbear — create a 1/1 red Goblin creature token that's tapped and attacking", "1/1 red Goblin", 1),
			},
			Text: "Whenever this creature attacks, create a 1/1 red Goblin creature token that's tapped and attacking.",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{R}: Until end of turn, this land becomes a 3/2 red Goblin creature with \"Whenever this creature attacks, create a 1/1 red Goblin creature token that's tapped and attacking.\" It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{3}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.AfrB17Manland(g, item, "Den of the Bugbear", "Goblin", "R", 3, 2, nil, denOfTheBugbearAttackToken)
			},
		}},
	})
}
