package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nimble Trapfinder — Creature — Human Rogue {1}{U}, 2/1 (Zendikar Rising,
// S82 OLD #5508):
//
//	"This creature can't be blocked if you had another Cleric, Rogue,
//	 Warrior, or Wizard enter the battlefield under your control this turn.
//	 At the beginning of combat on your turn, if you have a full party,
//	 creatures you control gain "Whenever this creature deals combat damage
//	 to a player, draw a card" until end of turn."
//
// The first line is a layer 6 restriction that holds while the turn's entry
// tally says another party creature entered (S82OldB089AnotherPartyTypeEnteredThisTurn).
// The combat trigger is an intervening "if" on a full party (CR 700.8c,
// 603.4), and grants the draw trigger to the creatures you control as it
// resolves (CR 611.2c); each carries its own trigger.
//
// No simplification.
func init() {
	const grant = "nimble-trapfinder/draw-on-damage"
	effects.Register(effects.Spec{
		OracleID:     "e95a9488-3f86-4598-9c6e-6b1b959f8570",
		Name:         "Nimble Trapfinder",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Whenever this creature deals combat damage to a player, draw a card", effects.Do(effects.DrawCards{N: 1})),
			},
			Text: "Whenever this creature deals combat damage to a player, draw a card.",
		}},
		Static: []game.StaticAbility{{
			Label: "This creature can't be blocked if you had another Cleric, Rogue, Warrior, or Wizard enter the battlefield under your control this turn.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.SelfOnly(target, g, source) && effects.S82OldB089AnotherPartyTypeEnteredThisTurn(g, source)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantBeBlocked
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.IfYouHaveAFullParty(effects.AtBeginningOfYourCombat("Nimble Trapfinder — creatures you control gain \"Whenever this creature deals combat damage to a player, draw a card\" until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GrantAbilitiesFor{
						Match: effects.And(effects.Creature(), effects.YouControl()),
						Keys:  []string{grant},
						Label: "Nimble Trapfinder — draw on combat damage",
					}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}
