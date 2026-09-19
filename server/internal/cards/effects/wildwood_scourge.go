package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wildwood Scourge — Creature — Hydra {X}{G}, 0/0 (EDHREC rank
// 3733):
//
//	"This creature enters with X +1/+1 counters on it.
//	 Whenever one or more +1/+1 counters are put on another non-Hydra
//	 creature you control, put a +1/+1 counter on this creature."
//
// The counters deck's snowball. The X counters are the printed entry
// clause (below), so the Scourge enters as an X/X. The growth trigger
// watches EventCounterPlaced: the engine emits one event per
// permanent per placement with the post-change total, so "one or
// more" is the event itself and the delta is read back off the log
// (b33CountersPlacedDelta — a removal is not a placement). The
// counters must land on a creature the controller controls that is
// neither the Scourge nor a Hydra (effective subtypes, so a
// changeling is a Hydra and does not count), and the Scourge grows
// only if it is still on the battlefield when the trigger resolves.
// Its own X counters are placed on itself, so they never fire it.
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
// A Scourge cast for X=0 is a printed 0/0 with no counters and dies
// to the toughness check, as in paper. That sentence was written here
// before it was true — the check read every printed 0/0 as the
// importer's stand-in and the Scourge stayed on the battlefield, the
// false comment #691 was filed to fix. #691 made the engine match the
// comment instead of the other way round: an object with a printing
// behind it has a real printed body (game.Card.ToughnessIsKnown), and
// printed_zero_body_test.go pins it.
func init() {
	Register(Spec{
		OracleID:                   "b9dec104-c636-4770-a7fc-7a3331face15",
		Name:                       "Wildwood Scourge",
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		XMatters:                   true,
		Completeness:               CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b35PlusCountersPutOnAnotherNonHydraCreatureYouControl(ev, source, g)
			}, "Wildwood Scourge — put a +1/+1 counter on this creature", b35PutCounterOnSelf),
		},
	})
}
