package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Poisonbelly Ogre — Creature — Ogre Warrior {4}{B}, 3/3 (Guildpact):
//
//	"Whenever another creature enters, its controller loses 1 life."
//
// Any creature entering under any player's control, the Ogre's controller's
// own included. "Its controller" is read as the creature entered, as it last
// existed if it has already left (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "38872730-0fd1-414e-8015-29d1bc649913",
		Name:         "Poisonbelly Ogre",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AnotherCreatureEntered,
				"Poisonbelly Ogre — its controller loses 1 life",
				func(g *game.Game, item *game.StackItem) error {
					info, ok := effects.NewContext(g, item).TriggeringPermanent()
					if !ok {
						return nil
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, info.Controller, -1)
				}),
		},
	})
}
