package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mardu Strike Leader — {2}{B} Creature — Human Warrior 3/2
//
//	"Whenever this creature attacks, create a 2/1 black Warrior creature token.
//	 Dash {3}{B}"
func init() {
	effects.Register(effects.Spec{
		OracleID:         "9be8e04f-f87f-4477-93f3-c546e9b346f0",
		Name:             "Mardu Strike Leader",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Dash("{3}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Mardu Strike Leader — create a 2/1 black Warrior creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/1 black Warrior"), N: 1})),
		},
	})
}
