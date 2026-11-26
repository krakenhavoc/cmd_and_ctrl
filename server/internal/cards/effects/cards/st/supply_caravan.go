package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Supply Caravan — {4}{W} Creature — Camel 3/5 (Amonkhet, #5535):
//
//	"When this creature enters, if you control a tapped creature, create a
//	 1/1 white Warrior creature token with vigilance."
//
// An intervening "if" (CR 603.4): the creature must be tapped as the
// ability would trigger and again as it resolves. Supply Caravan enters
// untapped, so it never satisfies its own clause.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1abcfe8d-4beb-4910-b353-1227e1256329",
		Name:         "Supply Caravan",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you control a tapped creature", func(g *game.Game, you uuid.UUID) bool {
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == you && c.IsCreature() && c.Tapped {
						return true
					}
				}
				return false
			}, effects.WhenThisEnters("Supply Caravan — create a 1/1 white Warrior with vigilance",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Warrior with vigilance"), N: 1}))),
		},
	})
}
