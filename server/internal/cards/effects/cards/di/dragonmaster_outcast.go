package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragonmaster Outcast — Creature — Human Shaman {R}, 1/1 (EDHREC
// rank 1540):
//
//	"At the beginning of your upkeep, if you control six or more
//	 lands, create a 5/5 red Dragon creature token with flying."
//
// A one-drop that makes a Dragon every turn once the land count is
// there. The "if you control six or more lands" is an intervening-if
// (CR 603.4): checked when the upkeep begins, so with five lands the
// ability never goes on the stack, and checked again as it resolves,
// so a land destroyed in response makes no Dragon. Lands are the
// post-layer type, so an animated or type-changed permanent counts
// exactly as it does on paper.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6fb79c3-cd32-4045-8177-e52841eea65b",
		Name:         "Dragonmaster Outcast",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && effects.B10LandsControlled(g, source.Controller) >= 6
			}, "Dragonmaster Outcast — create a 5/5 red Dragon with flying", func(g *game.Game, item *game.StackItem) error {
				if effects.B10LandsControlled(g, item.Controller) < 6 {
					return nil
				}
				return effects.CreateToken{Controller: item.Controller, Template: effects.B14RedDragonToken(5), N: 1}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
