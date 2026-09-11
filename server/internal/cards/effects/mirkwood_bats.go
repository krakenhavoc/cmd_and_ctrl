package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mirkwood Bats — Creature — Bat {3}{B}, 2/3 (EDHREC rank 227):
//
//	"Flying
//	 Whenever you create or sacrifice a token, each opponent loses
//	 1 life."
//
// The token deck's Blood Artist: a Treasure made and a Treasure
// cracked is two drains, and a wide token board sacrificed to an
// altar is a lethal one. Both halves are real, and neither targets:
//
//   - "Create" watches EventTokenCreated, which CreateTokenForEffect
//     emits once per token with the controller in Actor — so a
//     five-token Krenko activation is five triggers, as printed.
//   - "Sacrifice" watches EventSacrifice, emitted BEFORE the zone
//     move (see events.go), so the sacrificed permanent is still
//     findable and its Token type line still readable. A token
//     that dies to damage or a wrath is not sacrificed and does not
//     trigger; a token sacrificed as a mana-ability cost (a cracked
//     Treasure) does.
//
// Life loss, not damage: no prevention or damage doubling applies.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0636b6c3-0662-420a-b30d-f0a14e7c512d",
		Name:            "Mirkwood Bats",
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventTokenCreated, game.EventSacrifice},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				switch ev.Kind {
				case game.EventTokenCreated:
					return true
				case game.EventSacrifice:
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && IsToken(c)
				}
				return false
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Mirkwood Bats — each opponent loses 1 life",
					func(g *game.Game, item *game.StackItem) error {
						return eachOpponentLosesLife(g, item, 1)
					})
			},
		}},
	})
}
