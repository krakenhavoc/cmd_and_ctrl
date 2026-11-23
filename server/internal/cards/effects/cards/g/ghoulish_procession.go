package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghoulish Procession — Enchantment {1}{B} (Midnight Hunt):
//
//	"Whenever one or more nontoken creatures die, create a 2/2 black
//	 Zombie creature token with decayed. This ability triggers only once
//	 each turn."
//
// "One or more" is a batch (CR 603.2c): a wrath is one Zombie
// (OncePerBatch). Any player's nontoken creature counts. The second
// batch of the turn does not trigger at all (TriggersOnlyOnceEachTurn).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "39e20898-cf01-48dc-8972-7ac500c3fa79",
		Name:         "Ghoulish Procession",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.TriggersOnlyOnceEachTurn(effects.OncePerBatch(effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && !effects.IsToken(dead)
			}, "Ghoulish Procession — create a 2/2 black Zombie creature token with decayed",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie with decayed"), N: 1})))),
		},
	})
}
