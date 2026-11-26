package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bonds of Mortality — Enchantment {1}{G} (Oath of the Gatewatch, #5545):
//
//	"When this enchantment enters, draw a card.
//	 {G}: Creatures your opponents control lose hexproof and indestructible
//	 until end of turn."
//
// The strip is Shadowspear's turn-scoped layer 6 removal over the creatures
// your opponents control as the ability resolves (CR 611.2c): a creature that
// arrives afterwards keeps both.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d93ba2f2-bf9b-444b-adbb-3835662734e7",
		Name:         "Bonds of Mortality",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Bonds of Mortality — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}: Creatures your opponents control lose hexproof and indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    effects.ManaCost("{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.UntilEndOfTurn(effects.NewContext(g, item), uuid.Nil,
					effects.And(effects.Creature(), effects.OpponentControls()),
					"Bonds of Mortality — opponents' creatures lose hexproof and indestructible",
					game.RemoveKeywordsMod("hexproof", "indestructible"))
			},
		}},
	})
}
