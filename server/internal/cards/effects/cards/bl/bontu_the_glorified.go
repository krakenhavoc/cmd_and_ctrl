package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bontu the Glorified — {2}{B} Legendary Creature — God 4/6 (Amonkhet, #5535):
//
//	"Menace, indestructible
//	 Bontu can't attack or block unless a creature died under your control
//	 this turn.
//	 {1}{B}, Sacrifice another creature: Scry 1. Each opponent loses 1 life
//	 and you gain 1 life."
//
// The restriction is a layer 6 static, lifted as soon as the controller's
// turn tally shows a creature that died under their control (the tally Cult
// Conscript reads). The sacrificed creature is a cost, so it counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c2679c32-a23d-4953-bb48-b46580439ce2",
		Name:            "Bontu the Glorified",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace", "indestructible"},
		Static: []game.StaticAbility{{
			Label: "Bontu can't attack or block unless a creature died under your control this turn.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.SelfOnly(target, g, source) && g.TurnTallyFor(source.Controller).CreaturesDied == 0
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack | game.CantBlock
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice another creature: Scry 1. Each opponent loses 1 life and you gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Scry: 1, LifeGain: 1},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Effect:  effects.Do(effects.Scry{N: 1}, effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1}),
		}},
	})
}
