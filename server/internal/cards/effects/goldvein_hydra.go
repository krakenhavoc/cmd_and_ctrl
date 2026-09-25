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
// The X counters are the printed CR 614.1c entry clause and ride the
// CR 614 pipeline as one — XCounters, seeded onto the entry event off
// the resolving stack item (#1002). They land on the PERMANENT, after
// the move and before EventETB, so the dies trigger below reads a
// finished creature, Doubling Season and Hardened Scales apply, and a
// "whenever one or more counters are put on a permanent you control"
// payoff sees them too.
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
		OracleID:                   "2b62543f-a475-457a-a96b-b5d070383d3c",
		Name:                       "Goldvein Hydra",
		XMatters:                   true,
		Completeness:               CompletenessFull,
		PrintedKeywords:            []string{"vigilance", "trample", "haste"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			Key: "Goldvein Hydra — create tapped Treasures equal to its power",
			// The last-known power is a board read at trigger time
			// (ADR 0041 P9's fill-in Build), carried as item.Params.Amount.
			Build: func(_ game.Event, source *game.Card, lki game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Goldvein Hydra — create tapped Treasures equal to its power", nil)
				item.Params.Amount = b13LastKnownPower(g, source.InstanceID, lki)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b13CreateTappedTreasures(NewContext(g, item), item.Controller, item.Params.Amount)
			},
		}},
	})
}
