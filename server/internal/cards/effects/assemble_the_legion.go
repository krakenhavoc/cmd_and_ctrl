package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Assemble the Legion — Enchantment {3}{R}{W} (EDHREC rank 2212):
//
//	"At the beginning of your upkeep, put a muster counter on this
//	 enchantment. Then create a 1/1 red and white Soldier creature
//	 token with haste for each muster counter on this enchantment."
//
// The Boros inevitability engine: one Soldier the first upkeep, two
// the next, three after that. The upkeep trigger puts the counter
// on the Legion, then reads the count back OFF the Legion and makes
// that many Soldiers — so a Doubling Season sees the counter land
// and the Legion grows twice as fast, as printed. If the Legion has
// left the battlefield by the time the trigger resolves the counter
// cannot be placed, and the Soldiers are made for the muster
// counters it had when the trigger fired (CR 608.2h's last-known
// count, carried by the item since the LKI characteristic has no
// counters).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f81bef3-6ca0-4cf4-aa99-7b2813eeef04",
		Name:         "Assemble the Legion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginUpkeep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			},
			Key: "Assemble the Legion — add a muster counter, then muster the Soldiers",
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Assemble the Legion — add a muster counter, then muster the Soldiers")
				item.Params.Amount = source.Counters["muster"]
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				lastKnown := item.Params.Amount
				// #1432: a new object the card became is not
				// "this enchantment" — it gets no counter, and
				// the muster reads the one that triggered.
				if !b15OnBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
					return assembleTheLegionMuster(NewContext(g, item), lastKnown)
				}
				// #1290: the count that matters is what the
				// counter LANDED AT, not the pre-placement
				// value — a Doubling Season / Hardened Scales
				// board pauses the placement on a CR 616
				// prompt, and reading Counters["muster"] on
				// the next line would see the pre-placement
				// count.
				then := NewContext(g, item).ItemTail(assembleTheLegionMusterTail)
				then.N = lastKnown
				return g.AddCounterByKeyedForEffect(uuid.Nil, item.SourceCardID, "muster", 1, then)
			},
		}},
	})
}

// assembleTheLegionMuster creates a Soldier for each of n muster counters.
func assembleTheLegionMuster(ctx *Context, n int) error {
	if n <= 0 {
		return nil
	}
	return CreateToken{Controller: ctx.Controller(), Template: TokenCard("1/1 red and white Soldier with haste"), N: n}.Apply(ctx)
}

// assembleTheLegionMusterTail is the muster, read at what the counter landed
// at; the tail's N is the count the trigger saw, kept if the enchantment is
// gone by then. An on-disk key: never renamed, never reused.
var assembleTheLegionMusterTail = CardTail("assemble-the-legion/muster", func(ctx *Context, o game.TailOutcome) error {
	n := o.Tail.N
	if c, ok := ctx.Game.LookupCardForEffect(ctx.Item.SourceCardID); ok {
		n = c.Counters["muster"]
	}
	return assembleTheLegionMuster(ctx, n)
})
