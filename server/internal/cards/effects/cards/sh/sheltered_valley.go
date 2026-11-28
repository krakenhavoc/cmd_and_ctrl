package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sheltered Valley — Land (Alliances, S82 OLD batch 820):
//
//	"If this land would enter, instead sacrifice each other permanent named
//	 Sheltered Valley you control, then put this land onto the battlefield.
//	 At the beginning of your upkeep, if you control three or fewer lands, you
//	 gain 1 life.
//	 {T}: Add {C}."
//
// The sacrifice is part of the entry (AsEntersOthers, CR 614.1c): the others
// go before any event announces the new one, and the new one is not among
// "each other". The upkeep clause is an intervening "if" (CR 603.4).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd535fa3-6fd8-4227-97fd-3ef07cb0598d",
		Name:         "Sheltered Valley",
		Completeness: effects.CompletenessFull,
		AsEntersOthers: func(self *game.Card, ctx *effects.Context) error {
			var others []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.InstanceID != self.InstanceID && c.Controller == self.Controller && c.Effective().Name == "Sheltered Valley" &&
					!ctx.Game.ArrivingForEffect(c.InstanceID) {
					others = append(others, c.InstanceID)
				}
			}
			if len(others) == 0 {
				return nil
			}
			return ctx.Game.SacrificeAllKeyedForEffect(uuid.Nil, others, game.KeyedTail{})
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you control three or fewer lands", func(g *game.Game, you uuid.UUID) bool {
				return effects.CountControlled(g, you, game.Card.IsLand) <= 3
			}, effects.AtYourUpkeep("Sheltered Valley — you gain 1 life", effects.Do(effects.GainLife{Amount: 1}))),
		},
	})
}
