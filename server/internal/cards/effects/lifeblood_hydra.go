package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lifeblood Hydra — Creature — Hydra {X}{G}{G}{G}, 0/0 (EDHREC rank
// 3036):
//
//	"Trample
//	 This creature enters with X +1/+1 counters on it.
//	 When this creature dies, you gain life and draw cards equal to
//	 its power."
//
// The Hydra that pays out on death. Trample rides PrintedKeywords;
// the X counters are the printed entry clause (below); the death
// trigger reads the Hydra's
// last-known power — the harvester's LKI characteristic carries the
// layer-computed P/T and the +1/+1 counters are read back off the
// log (b13LastKnownPower, Conclave Mentor's shape), so a pumped or
// grown Hydra pays out for what it was when it died — and gains that
// much life, then draws that many, in printed order.
//
// The X counters are the printed CR 614.1c entry clause and ride the
// CR 614 pipeline as one — XCounters, seeded onto the entry event off
// the resolving stack item while the spell is still there (#1002).
// They land on the PERMANENT, after the move and before EventETB, so
// Doubling Season and Hardened Scales apply, the card's own enters
// trigger reads a finished creature, and a "whenever one or more
// counters are put on a permanent you control" payoff sees them —
// which it could not while they went onto a card still on the stack.
//
// Cast for X=0 the Hydra enters as the printed 0/0 it is and the next
// state-based check puts it into its owner's graveyard (CR 704.5f),
// paying out nothing, exactly as in paper. CR 601.2b allows the
// announcement; it simply does not survive it. That was an engine gap
// until #691 — the toughness check read every printed 0/0 as the
// importer's stand-in — and what closed it is the printing behind the
// object (game.Card.ToughnessIsKnown).
func init() {
	Register(Spec{
		OracleID:                   "b14d05c0-fe10-4079-a90e-0aea1a8fd375",
		Name:                       "Lifeblood Hydra",
		XMatters:                   true,
		Completeness:               CompletenessFull,
		PrintedKeywords:            []string{"trample"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, lki game.Characteristic, g *game.Game) *game.StackItem {
				power := b13LastKnownPower(g, source.InstanceID, lki)
				return game.NewTriggeredItem(source, "Lifeblood Hydra — gain life and draw cards equal to its power",
					func(g *game.Game, item *game.StackItem) error {
						return b28GainLifeAndDraw(g, item, power)
					})
			},
		}},
	})
}
