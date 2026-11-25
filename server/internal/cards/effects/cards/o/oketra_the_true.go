package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oketra the True — {3}{W} Legendary Creature — God 3/6 (Amonkhet, #5535):
//
//	"Double strike, indestructible
//	 Oketra can't attack or block unless you control at least three other
//	 creatures.
//	 {3}{W}: Create a 1/1 white Warrior creature token with vigilance."
//
// The restriction is a layer 6 static that applies while the controller
// has fewer than three other creatures (Oketra herself never counts).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "646e68c8-cc42-4d6c-8363-a18b10a10eae",
		Name:            "Oketra the True",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike", "indestructible"},
		Static: []game.StaticAbility{{
			Label: "Oketra can't attack or block unless you control at least three other creatures.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				if !effects.SelfOnly(target, g, source) {
					return false
				}
				others := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == source.Controller && c.IsCreature() && c.InstanceID != source.InstanceID {
						others++
					}
				}
				return others < 3
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack | game.CantBlock
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}: Create a 1/1 white Warrior creature token with vigilance.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{3}{W}"),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Warrior with vigilance"), N: 1}),
		}},
	})
}
