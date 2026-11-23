package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const cruelSomnophageOracleID = "997bdec5-f67b-4822-a3ba-c636e2685e8a"

// Cruel Somnophage // Can't Wake Up — Creature — Nightmare {1}{B},
// */*, with an Adventure, Can't Wake Up {1}{U} (Sorcery):
//
//	"Cruel Somnophage's power and toughness are each equal to the
//	 number of creature cards in all graveyards."
//	"Target player mills four cards."
//
// The creature is Mortivore's characteristic-defining ability; the
// Adventure half is a registered "#1" entry like Bonecrusher Giant's
// Stomp, and the engine handles the exile and the later cast from it.
func init() {
	effects.Register(effects.Spec{
		OracleID:     cruelSomnophageOracleID,
		Name:         "Cruel Somnophage",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := effects.B41CreatureCardsInAllGraveyards(g)
				c.Power = n
				c.Toughness = n
			},
		}},
	})

	effects.Register(effects.Spec{
		OracleID:     cruelSomnophageOracleID + "#1",
		Name:         "Can't Wake Up",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			victim, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			return effects.MillCards{Player: victim, N: 4}.Apply(ctx)
		},
	})
}
