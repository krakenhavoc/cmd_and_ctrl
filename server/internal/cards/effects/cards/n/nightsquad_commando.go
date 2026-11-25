package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nightsquad Commando — Creature — Human Soldier {2}{B}, 2/3 (Ikoria: Lair
// of Behemoths, #5512):
//
//	"When this creature enters, if you attacked this turn, create a 1/1
//	 white Human Soldier creature token."
//
// An intervening "if" (CR 603.4), checked as it would trigger and again as
// it resolves, like Fire Nation Raider's raid trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "04cffc41-05da-45fb-82da-2c4d4108c7e1",
		Name:         "Nightsquad Commando",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.TmcB6YouAttackedThisTurn),
				"Nightsquad Commando — create a 1/1 white Human Soldier creature token",
				effects.Do(effects.If{Cond: effects.AvatarB11CondYouAttacked{}, Then: []effects.Applier{
					effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 1},
				}})),
		},
	})
}
