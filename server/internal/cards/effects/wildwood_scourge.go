package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wildwood Scourge — Creature — Hydra {X}{G}, 0/0 (EDHREC rank
// 3733):
//
//	"This creature enters with X +1/+1 counters on it.
//	 Whenever one or more +1/+1 counters are put on another non-Hydra
//	 creature you control, put a +1/+1 counter on this creature."
//
// The counters deck's snowball. The X counters go on as the spell
// resolves (Goldvein Hydra's posture, declared below), so the
// Scourge enters as an X/X. The growth trigger watches
// EventCounterPlaced: the engine emits one event per permanent per
// placement with the post-change total, so "one or more" is the
// event itself and the delta is read back off the log
// (b33CountersPlacedDelta — a removal is not a placement). The
// counters must land on a creature the controller controls that is
// neither the Scourge nor a Hydra (effective subtypes, so a
// changeling is a Hydra and does not count), and the Scourge grows
// only if it is still on the battlefield when the trigger resolves.
// Its own X counters are placed on itself, so they never fire it.
//
// One declared simplification, weaker than printed: the X +1/+1
// counters are placed as the spell resolves, a beat before the card
// enters (an entry replacement cannot read the spell's X), so a
// "whenever you put counters on a permanent" payoff does not see
// them. A Scourge cast for X=0 is a printed 0/0 with no counters and
// dies to the toughness check, as in paper.
func init() {
	Register(Spec{
		OracleID:     "b9dec104-c636-4770-a7fc-7a3331face15",
		Name:         "Wildwood Scourge",
		XMatters:     true,
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The X +1/+1 counters are put on the Scourge as the spell resolves, a beat before it enters, so effects that watch you put counters on a permanent don't see them."},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: ctx.X()}.Apply(ctx)
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b35PlusCountersPutOnAnotherNonHydraCreatureYouControl(ev, source, g)
			}, "Wildwood Scourge — put a +1/+1 counter on this creature", b35PutCounterOnSelf),
		},
	})
}
