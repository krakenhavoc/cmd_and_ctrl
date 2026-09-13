package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "b6fb79c3-cd32-4045-8177-e52841eea65b",
		Name:         "Dragonmaster Outcast",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginUpkeep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && b10LandsControlled(g, source.Controller) >= 6
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Dragonmaster Outcast — create a 5/5 red Dragon with flying",
					func(g *game.Game, item *game.StackItem) error {
						if b10LandsControlled(g, item.Controller) < 6 {
							return nil
						}
						return CreateToken{Controller: item.Controller, Template: b14RedDragonToken(5), N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
