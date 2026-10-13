package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// choose_counters.go — "remove up to three counters from target permanent"
// (Glissa Sunslayer) and "move any number of counters from target
// permanent onto another" (Resourceful Defense): a number of counters of
// the player's choice, of whatever kinds the permanent carries, chosen as
// the ability resolves (CR 608.2c).
//
// It is ChooseNumber asked once per kind of counter on the permanent, in a
// fixed order (sorted by kind), each bounded by what the permanent has and
// by what is left of the total ceiling. A permanent with one kind of
// counter is one question; a planeswalker carrying loyalty and a stun
// counter is two. The answer is the same thing printed text means by "any
// number" of counters, without a new prompt kind.
//
// The counts are read when the question is asked. The effect re-reads the
// permanent when it acts on the answer, so a counter that left in between
// is simply not removed (AddCounter clamps at zero).

// counterGoal is what a bot should pick for one kind of counter: `most` is
// the most it may choose.
type counterGoal func(ctx *Context, kind string, most int) int

// chooseCounters asks, kind by kind, how many counters to choose on `perm`
// — at most `ceiling` in all, or any number when ceiling is 0 — and calls
// `then` once with the counts (kind to count, zero counts dropped).
// Nothing to choose from calls `then` with an empty map.
func chooseCounters(ctx *Context, perm uuid.UUID, ceiling int, question string, goal counterGoal, then func(ctx *Context, picked map[string]int) error) error {
	c, ok := ctx.Game.LookupCardForEffect(perm)
	if !ok {
		return then(ctx, map[string]int{})
	}
	var kinds []string
	for kind, n := range c.Counters {
		if n > 0 {
			kinds = append(kinds, kind)
		}
	}
	sort.Strings(kinds)
	have := map[string]int{}
	for _, kind := range kinds {
		have[kind] = c.Counters[kind]
	}
	return askCounterKind(ctx, kinds, have, ceiling, ceiling > 0, map[string]int{}, question, goal, then)
}

func askCounterKind(ctx *Context, kinds []string, have map[string]int, left int, limited bool, picked map[string]int, question string, goal counterGoal, then func(ctx *Context, picked map[string]int) error) error {
	for len(kinds) > 0 {
		kind := kinds[0]
		most := have[kind]
		if limited && most > left {
			most = left
		}
		if most <= 0 {
			kinds = kinds[1:]
			continue
		}
		rest := kinds[1:]
		return ChooseNumber{
			Min:      0,
			Max:      most,
			Question: question + " — " + kind + " counters",
			Unit:     game.PayAmountCounters,
			Goal: func(ctx *Context) int {
				if goal == nil {
					return most
				}
				n := goal(ctx, kind, most)
				if n < 0 {
					return 0
				}
				if n > most {
					return most
				}
				return n
			},
			Then: func(ctx *Context, n int) error {
				if n > 0 {
					picked[kind] += n
				}
				nextLeft := left
				if limited {
					nextLeft -= n
				}
				return askCounterKind(ctx, rest, have, nextLeft, limited, picked, question, goal, then)
			},
		}.Apply(ctx)
	}
	return then(ctx, picked)
}
