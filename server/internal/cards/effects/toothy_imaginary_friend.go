package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Toothy, Imaginary Friend — Legendary Creature — Illusion {3}{U}, 1/1:
//
//	"Partner with Pir, Imaginative Rascal (When this creature enters,
//	 target player may put Pir into their hand from their library,
//	 then shuffle.)
//	 Whenever you draw a card, put a +1/+1 counter on Toothy.
//	 When Toothy leaves the battlefield, draw a card for each +1/+1
//	 counter on it."
//
// Partner with is CR 702.124j's two abilities: the commander pairing
// with Pir (internal/deck) and the entry search (PartnerWith, #2142).
//
// The draw trigger fires once per card drawn, as Chasm Skulker's does.
// The leaves trigger counts the counters Toothy had as it left
// (last-known information, CR 608.2h), since by the time it resolves
// Toothy is in another zone with none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "41d6cce0-b852-4d0e-aee2-081df13dd9b8",
		Name:         "Toothy, Imaginary Friend",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			PartnerWith("Toothy, Imaginary Friend", "Pir, Imaginative Rascal"),
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			}, "Toothy, Imaginary Friend — put a +1/+1 counter on Toothy",
				func(g *game.Game, item *game.StackItem) error {
					return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
				}),
			WhenThisLeaves("Toothy, Imaginary Friend — draw a card for each +1/+1 counter on it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					info, ok := ctx.SourcePermanent()
					if !ok {
						return nil
					}
					return DrawCards{Player: item.Controller, N: info.Counters[game.CounterPlusOne]}.Apply(ctx)
				}),
		},
	})
}
