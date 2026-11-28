package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aether Flash — {2}{R}{R} Enchantment (Weatherlight, S82 OLD batch 771, #5629):
//
//	"Whenever a creature enters, this enchantment deals 2 damage to it."
//
// One trigger per creature that enters the battlefield, whoever controls it
// (Monkey Cage's entering-creature condition); the damage goes to that creature
// if it is still on the battlefield as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3c35742-e306-49b6-b042-db4f685c6f86",
		Name:         "Aether Flash",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B33AnyCreatureEntered(ev, g)
			}, "Aether Flash — it deals 2 damage to that creature", effects.S82OldB771DamageTheEnteredCreature),
		},
	})
}
