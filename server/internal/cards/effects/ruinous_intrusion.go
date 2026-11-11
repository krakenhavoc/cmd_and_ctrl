package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ruinous Intrusion — Instant {3}{G} (Lost Caverns of Ixalan Commander,
// #5368):
//
//	"Exile target artifact or enchantment. Put X +1/+1 counters on target
//	 creature you control, where X is the mana value of the permanent
//	 exiled this way."
//
// Two target clauses. X is the exiled permanent's mana value as it was on
// the battlefield (a token or copy counts as its copied values), read
// before the exile and handed to the tail that runs once the card has
// really reached exile: an artifact that was countered by a replacement
// or never exiled (illegal target) gives no X, so no counters. The
// creature is re-checked as it resolves (CR 608.2b); if only it is
// illegal the exile still happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4f945baa-40a9-4494-b2f9-f1926f3fec84",
		Name:         "Ruinous Intrusion",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: Clauses(
			TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			TargetCreature("target creature you control", YouControl()),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			victim, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(victim.ID)
			if !ok {
				return nil
			}
			tail := ctx.ItemTail(cmdr2B23RuinousIntrusionExiled)
			tail.N = c.ManaValue()
			return ExileTarget{Target: victim.ID, ThenTail: tail}.Apply(ctx)
		},
	})
}

// cmdr2B23RuinousIntrusionExiled is the second sentence: X +1/+1 counters
// on the creature chosen for the second clause. An on-disk key: never
// renamed, never reused.
var cmdr2B23RuinousIntrusionExiled = CardTail("ruinous-intrusion/after-exile", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || o.Tail.N <= 0 {
		return nil
	}
	mine, ok := ctx.ClauseTarget(1)
	if !ok {
		return nil
	}
	return AddCounter{Target: mine.ID, Kind: game.CounterPlusOne, N: o.Tail.N}.Apply(ctx)
})
