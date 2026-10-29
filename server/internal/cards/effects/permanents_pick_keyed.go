package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// permanents_pick_keyed.go — ChoosePermanents with its set rule, its
// sacrifice and the rest of the card REGISTERED rather than held in
// closures (#4290), so a table waiting on any seat's pick is a restore
// point.
//
// A card registers its pick once, from a package-level var, under an
// on-disk key (ledgered "run <key>": never renamed, never reused), and
// passes it as ChoosePermanents.Keyed. What the closure captured is data on
// the run: the resolving item itself, which the body's Context is rebuilt
// from, Carry (IDs) and Nums.

// PermanentsDef is the registered part of a ChoosePermanents.
type PermanentsDef struct {
	// Validate is the rule over the SET one seat picks (Tragic Arrogance's
	// "an artifact, a creature, an enchantment, and a planeswalker"),
	// applied to every leg. It is a function of the cards alone, because
	// it runs under the enumerator's read lock; it captures nothing.
	Validate func(picked []game.Card) bool

	// Sacrifice sacrifices everything chosen, across every leg, as one
	// simultaneous exit before Then runs (ChoosePermanents.Sacrifice).
	Sacrifice bool

	// Then is everything printed after the choice, run once with what
	// every seat picked. Its Context is rebuilt from the resolving item.
	// ThenRun is Then for a body that also reads the run's Carry or Nums;
	// set at most one.
	Then    func(ctx *Context, picked game.PromptedPicks) error
	ThenRun func(ctx *Context, picked game.PromptedPicks, r game.RunDone) error
}

// PermanentsThen names a registered permanent pick. Hold one only from
// RegisterPermanentsPick.
type PermanentsThen struct{ run game.PromptRunThen }

// RegisterPermanentsPick registers a permanent pick under key. Call it
// once, from a package-level var, or from init when the body asks the
// same pick again.
func RegisterPermanentsPick(key string, def PermanentsDef) PermanentsThen {
	body := func(ctx *Context, picked game.PromptedPicks, r game.RunDone) error {
		switch {
		case def.ThenRun != nil:
			return def.ThenRun(ctx, picked, r)
		case def.Then != nil:
			return def.Then(ctx, picked)
		}
		return nil
	}
	then := PermanentsThen{run: RunThenRule(key, def.Validate, func(ctx *Context, r game.RunDone) error {
		picked := r.Picks()
		if !def.Sacrifice || picked.Count() == 0 {
			return body(ctx, picked, r)
		}
		// #4253: the rest of the card is the sacrifice's keyed tail, naming
		// this pick by its key, with the run's answer and data on it.
		return ctx.Game.SacrificeAllKeyedForEffect(uuid.Nil, picked.Cards(), permanentsSacrificedTail(key, r))
	})}
	if def.Sacrifice {
		permanentsPickBodies[key] = body
	}
	return then
}

// permanentsPickBodies are the registered picks that sacrifice what was
// chosen, by key: the body permanentsPickSacrificed runs once the sacrifice
// has settled. Written by package initialisation only.
var permanentsPickBodies = map[string]func(ctx *Context, picked game.PromptedPicks, r game.RunDone) error{}

// permanentsSacrificedTail is the keyed rest of a sacrificing pick: the
// pick's key as Word, the run's asking item, and the run's carry, numbers
// and answer packed into Carry and Nums (encodeRunDone).
func permanentsSacrificedTail(key string, r game.RunDone) game.KeyedTail {
	carry, nums := encodeRunDone(r)
	return game.KeyedTail{Ref: permanentsPickSacrificed, Source: r.Source, Word: key, Asked: r.Asked, Carry: carry, Nums: nums}
}

// permanentsPickSacrificed runs a sacrificing pick's body (#4253) once its
// sacrifice has settled, against a Context rebuilt from the run's asking
// item, handed the run as it was when the last seat answered. An on-disk
// key.
var permanentsPickSacrificed = game.RegisterTail("permanents-pick/sacrificed", func(g *game.Game, o game.TailOutcome) error {
	body, ok := permanentsPickBodies[o.Tail.Word]
	if !ok {
		return nil
	}
	r := decodeRunDone(o.Tail)
	return body(runContext(g, r), r.Picks(), r)
})

// encodeRunDone packs a run's data and answer into a tail's Carry and Nums:
// Carry is the run's carry, then each seat followed by its cards; Nums is
// the carry's length, the numbers' count, the numbers, then each seat's
// card count.
func encodeRunDone(r game.RunDone) ([]uuid.UUID, []int) {
	carry := append([]uuid.UUID(nil), r.Carry...)
	nums := append([]int{len(r.Carry), len(r.Nums)}, r.Nums...)
	for _, sc := range r.Landed {
		carry = append(append(carry, sc.Seat), sc.Cards...)
		nums = append(nums, len(sc.Cards))
	}
	return carry, nums
}

// decodeRunDone is encodeRunDone's inverse, over the tail it was written
// on. A tail whose numbers do not add up is read as far as they do.
func decodeRunDone(t game.KeyedTail) game.RunDone {
	r := game.RunDone{Source: t.Source, Asked: t.Asked}
	carry, nums := t.Carry, t.Nums
	if len(nums) < 2 || nums[0] > len(carry) || 2+nums[1] > len(nums) {
		return r
	}
	r.Carry = append([]uuid.UUID(nil), carry[:nums[0]]...)
	r.Nums = append([]int(nil), nums[2:2+nums[1]]...)
	carry, nums = carry[nums[0]:], nums[2+nums[1]:]
	for _, n := range nums {
		if n < 0 || 1+n > len(carry) {
			break
		}
		r.Landed = append(r.Landed, game.SeatCards{Seat: carry[0], Cards: append([]uuid.UUID(nil), carry[1:1+n]...)})
		carry = carry[1+n:]
	}
	return r
}
