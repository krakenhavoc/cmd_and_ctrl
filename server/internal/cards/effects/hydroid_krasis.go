package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hydroid Krasis — Creature — Jellyfish Hydra Beast {X}{G}{U}, 0/0
// (EDHREC rank 1536):
//
//	"When you cast this spell, you gain half X life and draw half X
//	 cards. Round down each time.
//	 Flying, trample
//	 This creature enters with X +1/+1 counters on it."
//
// The Simic X-drop whose payout cannot be countered. The cast trigger
// is a FromStack ability (cascade's mechanism): it fires when the
// spell is announced, goes on the stack above it, and resolves first
// — so a Counterspell on the Krasis still leaves the life and the
// cards behind, as printed. X is read off the stack item at trigger
// time and halved, rounding down, each half separately.
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
// Cast for X=0 the Krasis enters as the printed 0/0 it is and the
// next state-based check puts it into its owner's graveyard (CR
// 704.5f), as in paper. CR 601.2b allows the announcement; it simply
// does not survive it. That was an engine gap until #691 — the
// toughness check read every printed 0/0 as the importer's stand-in —
// and what closed it is the printing behind the object
// (game.Card.ToughnessIsKnown).
func init() {
	Register(Spec{
		OracleID:                   "6bd872b2-5c40-4e11-9a7f-0136a51b0642",
		Name:                       "Hydroid Krasis",
		XMatters:                   true,
		Completeness:               CompletenessFull,
		PrintedKeywords:            []string{"flying", "trample"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		Triggered: []game.TriggeredAbility{{
			FromStack: true,
			Watches:   []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				half := 0
				if item := g.StackItemForEffect(source.InstanceID); item != nil && item.XValue > 0 {
					half = item.XValue / 2
				}
				return &game.StackItem{
					Kind:         game.StackItemTriggered,
					Controller:   ev.Actor,
					Owner:        ev.Actor,
					SourceCardID: source.InstanceID,
					Label:        "Hydroid Krasis — gain half X life and draw half X cards",
					Effect: func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						if err := (GainLife{Player: item.Controller, Amount: half}).Apply(ctx); err != nil {
							return err
						}
						return DrawCards{Player: item.Controller, N: half}.Apply(ctx)
					},
				}
			},
		}},
	})
}
