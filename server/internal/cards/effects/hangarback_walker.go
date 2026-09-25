package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hangarback Walker — Artifact Creature — Construct {X}{X}, 0/0
// (EDHREC rank 1528):
//
//	"This creature enters with X +1/+1 counters on it.
//	 When this creature dies, create a 1/1 colorless Thopter artifact
//	 creature token with flying for each +1/+1 counter on this
//	 creature.
//	 {1}, {T}: Put a +1/+1 counter on this creature."
//
// The artifact that turns into Thopters. Three abilities, all live:
// the X-sized body, the tap-to-grow, and the death payout — Thopters
// equal to the +1/+1 counters it had when it died.
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
// "For each +1/+1 counter on this creature" on death is last-known
// information (CR 603.10), read at RESOLUTION through
// ctx.TriggeringPermanent() (#1379, CR 608.2h): counters included, and
// keyed by the object's epoch, so a reanimation in response before the
// trigger resolves is a new object and cannot leak new counters into
// this reading — the count is the one the creature had when it died, a
// Corpsejack doubling included.
//
// Cast for X=0 the Walker enters as the printed 0/0 it is and the
// next state-based check puts it into its owner's graveyard (CR
// 704.5f) before the tap ability can grow it, making no Thopters,
// exactly as in paper. CR 601.2b allows the announcement; it simply
// does not survive it. That was an engine gap until #691 — the
// toughness check read every printed 0/0 as the importer's stand-in —
// and what closed it is the printing behind the object
// (game.Card.ToughnessIsKnown).
func init() {
	Register(Spec{
		OracleID:                   "dde55256-5259-44e7-a267-fca45a7f0d04",
		Name:                       "Hangarback Walker",
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		XMatters:                   true,
		Completeness:               CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			Key: "Hangarback Walker — create a Thopter for each +1/+1 counter it had",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				info, _ := ctx.TriggeringPermanent()
				n := info.Counters["+1/+1"]
				if n <= 0 {
					return nil
				}
				return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: n}.Apply(ctx)
			},
		}},
		Activated: []ActivatedAbility{{
			Label: "{1}, {T}: Put a +1/+1 counter on this creature.",
			Cost:  Plus(ManaCost("{1}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneBattlefield {
					return nil
				}
				return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
