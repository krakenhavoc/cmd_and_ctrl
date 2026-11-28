package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dingus Staff — {4} Artifact (Weatherlight, S82 OLD batch 771, #5629):
//
//	"Whenever a creature dies, this artifact deals 2 damage to that creature's
//	 controller."
//
// One trigger per creature that dies, any player's, including the one that owns
// the Staff. The damage goes to the controller the creature had as it last
// existed on the battlefield (CR 603.10a), a player who has since left the game
// excepted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7127fc35-600c-4397-8c01-5327d3c4944b",
		Name:         "Dingus Staff",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := effects.DiedCreature(ev, g)
				return ok
			}, "Dingus Staff — it deals 2 damage to that creature's controller", effects.S82OldB771DamageTheDiedCreaturesController),
		},
	})
}
