package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Starwinder — Creature — Leviathan {5}{U}{U}, 7/7 (Edge of Eternities,
// #4461):
//
//	"Whenever a creature you control deals combat damage to a player, you
//	 may draw that many cards.
//	 Warp {2}{U}{U}"
//
// Enduring Curiosity's watch (combat damage to a player by a creature you
// control, one trigger per creature) with Cold-Eyed Selkie's optional
// "draw that many": the count is the damage that creature dealt, and the
// question is asked as the trigger resolves (CR 603.5). Warp is the
// shared alternative cost (CR 702.185).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "face2dc0-37ba-4efe-b900-b5521567c943",
		Name:             "Starwinder",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Warp("{2}{U}{U}")},
		Triggered: []game.TriggeredAbility{
			OptionalDrawThatMany(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			}, "Starwinder — draw that many cards", func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: coldEyedSelkieDraws(g, item)}.Apply(NewContext(g, item))
			}), "Starwinder — draw that many cards?", coldEyedSelkieDraws),
		},
	})
}
