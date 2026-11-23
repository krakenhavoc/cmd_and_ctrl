package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nested Ghoul — {3}{B}{B} Creature — Phyrexian Zombie Warrior 4/2 (Jumpstart 2022, #5948):
//
//	"Whenever a source deals damage to this creature, create a 2/2 black
//	 Phyrexian Zombie creature token."
//
// One token per damage event that dealt damage (a damage event with an
// amount names damage actually dealt, after prevention).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e60b6b71-eea8-42b3-81e2-c8fb1af5e218",
		Name:         "Nested Ghoul",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventDealDamage && ev.Target == source.InstanceID && ev.Amount > 0
			}, "Nested Ghoul — create a 2/2 black Phyrexian Zombie creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Phyrexian Zombie"), N: 1})),
		},
	})
}
