package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goldvein Hydra — Creature — Hydra {X}{G}, 0/0 (EDHREC rank 1448):
//
//	"Vigilance, trample, haste
//	 This creature enters with X +1/+1 counters on it.
//	 When this creature dies, create a number of tapped Treasure
//	 tokens equal to its power."
//
// The X-drop that refunds itself. Three printed keywords, an X-sized
// body, and a dies trigger that reads the body's size.
//
// Sandbox simplification, declared, for the X counters: an entry
// replacement cannot see the X announced for the spell (the stack
// item is gone by the time the entry pipeline runs), so the counters
// go on as the spell RESOLVES — a beat before the card moves from
// the stack to the battlefield, which is the last moment X is
// readable — rather than as an "enters with" replacement. They are
// on the card when it lands, so the 0/0 body never meets the
// state-based check without them (X=0 dies at once, as printed),
// every ETB watcher sees the finished creature, and a counter
// doubler or Hardened Scales applies, exactly as it does to a
// printed "enters with". The one observable difference is that a
// "whenever you put counters on a permanent" payoff does not see
// them, because the card was not a permanent yet — weaker, never
// stronger.
//
// "Its power" on death is the CR 603.10 last-known power. The LKI
// characteristic the harvester hands the trigger carries the layer
// math but not the counters (the card's Counters are cleared by the
// move), so the +1/+1 and -1/-1 totals are read back off the event
// log — b13LastKnownPower. Tapped Treasures, as printed.
//
// Cast for X=0 the Hydra enters as the printed 0/0 it is and the
// next state-based check puts it into its owner's graveyard (CR
// 704.5f), making no Treasures, exactly as in paper. CR 601.2b allows
// the announcement; it simply does not survive it. That was an engine
// gap until #691 — the toughness check read every printed 0/0 as the
// importer's stand-in — and what closed it is the printing behind the
// object (game.Card.ToughnessIsKnown).
func init() {
	Register(Spec{
		OracleID:        "2b62543f-a475-457a-a96b-b5d070383d3c",
		Name:            "Goldvein Hydra",
		XMatters:        true,
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The X +1/+1 counters are put on the Hydra as the spell resolves, a beat before it enters, so effects that watch you put counters on a permanent don't see them."},
		PrintedKeywords: []string{"vigilance", "trample", "haste"},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: ctx.X()}.Apply(ctx)
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, lki game.Characteristic, g *game.Game) *game.StackItem {
				power := b13LastKnownPower(g, source.InstanceID, lki)
				return game.NewTriggeredItem(source, "Goldvein Hydra — create tapped Treasures equal to its power",
					func(g *game.Game, item *game.StackItem) error {
						return b13CreateTappedTreasures(NewContext(g, item), item.Controller, power)
					})
			},
		}},
	})
}
