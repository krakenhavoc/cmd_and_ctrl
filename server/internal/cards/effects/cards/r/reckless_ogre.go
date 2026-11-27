package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reckless Ogre — Creature — Ogre {3}{R}, 3/2 (Exodus, S82 OLD b740, #5626):
//
//	"Whenever this creature attacks alone, it gets +3/+0 until end of turn."
//
// Rogue Kavu's attack trigger with +3/+0: it fires only when this creature is
// the only creature declared as an attacker.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fe455177-d9e5-4893-ad87-ebbba5e2fa45",
		Name:         "Reckless Ogre",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, c game.Characteristic, g *game.Game) bool {
				return effects.ThisAttacked(ev, source, c, g) && game.AttackedAlone(g)
			}, "Reckless Ogre — it gets +3/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(3, 0, "Reckless Ogre — +3/+0 until end of turn")),
		},
	})
}
