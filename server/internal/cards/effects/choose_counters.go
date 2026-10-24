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
//
// A clause is REGISTERED (RegisterCounterChoice, #4294): its bot goal and
// what it does with the counts are package-level functions under one key,
// and the chain's place (which kinds, how many of each are on the
// permanent and chosen so far, how many may still be chosen) rides the
// prompt as data, so a table waiting on any of the questions is a restore
// point.

// counterGoal is what a bot should pick for one kind of counter on
// `perm`: `most` is the most it may choose.
type counterGoal func(ctx *Context, perm uuid.UUID, kind string, most int) int

// counterChoice is a registered "choose counters" clause.
type counterChoice struct {
	then  game.NumberThen
	goal  counterGoal
	final func(ctx *Context, perm uuid.UUID, extra []uuid.UUID, picked map[string]int) error
}

// registerCounterChoice registers a clause under key (a number branch,
// ledgered "number <key>"): `goal` is the bot's reading of one kind
// (nil: as many as it may), `final` what the clause does with the counts
// (kind to count, zero counts dropped) once every kind has been asked.
// Both capture nothing; the card's own IDs ride as `extra`.
func registerCounterChoice(key string, goal counterGoal, final func(ctx *Context, perm uuid.UUID, extra []uuid.UUID, picked map[string]int) error) *counterChoice {
	cc := &counterChoice{goal: goal, final: final}
	cc.then = NumberThen(key, func(ctx *Context, r game.NumberAnswer) error {
		s, ok := counterAskOf(r)
		if !ok {
			return nil
		}
		if r.Number > 0 {
			s.picked[s.idx] += r.Number
			if s.limited {
				s.left -= r.Number
			}
		}
		s.idx++
		return cc.ask(ctx, s)
	})
	return cc
}

// counterAsk is a chain of counter questions, where it is. It rides the
// prompt as the Carry (the permanent, then the clause's own IDs), the
// Words (the question, then the kinds in asking order) and the Nums (what
// is left to choose, whether the total is limited, the index of the kind
// being asked, then each kind's count on the permanent and each kind's
// count chosen so far).
type counterAsk struct {
	perm     uuid.UUID
	extra    []uuid.UUID
	question string
	kinds    []string
	have     []int
	picked   []int
	idx      int
	left     int
	limited  bool
}

func (s counterAsk) carry() []uuid.UUID {
	return append([]uuid.UUID{s.perm}, s.extra...)
}

func (s counterAsk) words() []string {
	return append([]string{s.question}, s.kinds...)
}

func (s counterAsk) nums() []int {
	limited := 0
	if s.limited {
		limited = 1
	}
	out := []int{s.left, limited, s.idx}
	out = append(out, s.have...)
	return append(out, s.picked...)
}

// counterAskOf reads the chain back off an answer. False for data that is
// not a chain's (a newer or older shape), which does nothing.
func counterAskOf(r game.NumberAnswer) (counterAsk, bool) {
	if len(r.Carry) == 0 || len(r.Words) == 0 || len(r.Nums) < 3 {
		return counterAsk{}, false
	}
	m := len(r.Words) - 1
	if len(r.Nums) != 3+2*m {
		return counterAsk{}, false
	}
	s := counterAsk{
		perm:     r.Carry[0],
		extra:    append([]uuid.UUID(nil), r.Carry[1:]...),
		question: r.Words[0],
		kinds:    append([]string(nil), r.Words[1:]...),
		left:     r.Nums[0],
		limited:  r.Nums[1] != 0,
		idx:      r.Nums[2],
		have:     append([]int(nil), r.Nums[3:3+m]...),
		picked:   append([]int(nil), r.Nums[3+m:]...),
	}
	return s, s.idx >= 0
}

// choose asks, kind by kind, how many counters to choose on `perm` — at
// most `ceiling` in all, or any number when ceiling is 0 — and runs the
// clause's final step once with the counts. Nothing to choose from runs
// it with an empty map.
func (cc *counterChoice) choose(ctx *Context, perm uuid.UUID, extra []uuid.UUID, ceiling int, question string) error {
	c, ok := ctx.Game.LookupCardForEffect(perm)
	if !ok {
		return cc.final(ctx, perm, extra, map[string]int{})
	}
	var kinds []string
	for kind, n := range c.Counters {
		if n > 0 {
			kinds = append(kinds, kind)
		}
	}
	sort.Strings(kinds)
	have := make([]int, len(kinds))
	for i, kind := range kinds {
		have[i] = c.Counters[kind]
	}
	return cc.ask(ctx, counterAsk{
		perm: perm, extra: extra, question: question, kinds: kinds, have: have,
		picked: make([]int, len(kinds)), left: ceiling, limited: ceiling > 0,
	})
}

func (cc *counterChoice) ask(ctx *Context, s counterAsk) error {
	for ; s.idx < len(s.kinds); s.idx++ {
		kind, most := s.kinds[s.idx], s.have[s.idx]
		if s.limited && most > s.left {
			most = s.left
		}
		if most <= 0 {
			continue
		}
		perm := s.perm
		return ChooseNumber{
			Min:      0,
			Max:      most,
			Question: s.question + " — " + kind + " counters",
			Unit:     game.PayAmountCounters,
			Goal: func(ctx *Context) int {
				if cc.goal == nil {
					return most
				}
				n := cc.goal(ctx, perm, kind, most)
				if n < 0 {
					return 0
				}
				if n > most {
					return most
				}
				return n
			},
			Then:  cc.then,
			Carry: s.carry(),
			Words: s.words(),
			Nums:  s.nums(),
		}.Apply(ctx)
	}
	picked := map[string]int{}
	for i, kind := range s.kinds {
		if s.picked[i] > 0 {
			picked[kind] = s.picked[i]
		}
	}
	return cc.final(ctx, s.perm, s.extra, picked)
}
