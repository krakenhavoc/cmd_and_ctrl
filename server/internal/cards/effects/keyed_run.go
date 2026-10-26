package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyed_run.go — a prompted run (a sacrifice, a choose_cards fan-out, a
// permanent pick) whose rest of the card is a registered continuation and
// plain data instead of a closure (#4290, ADR 0013 §5x / §5y), so a table
// waiting on any seat's leg is a restore point.
//
// A card registers its continuation once, from a package-level var, under
// an on-disk key (ledgered "run <key>": never renamed, never reused), and
// queues the run with the data the continuation reads: Carry (IDs), Nums,
// and the asking item itself, which every Context.RunData carries. The
// continuation is handed the answer (game.RunDone) and a Context rebuilt
// from the asking item (AskedItem.StackItem), so a body that read the
// item's targets, its X or the words of its text reads them as it did.

// RunThen registers the keyed rest of a card after a prompted run. The
// body runs against a Context rebuilt from the run's asking item, so a run
// queued with Context.RunData reads the item it came from. Call it once,
// from a package-level var.
func RunThen(key string, body func(ctx *Context, r game.RunDone) error) game.PromptRunThen {
	return game.RegisterPromptRunThen(key, func(g *game.Game, r game.RunDone) error {
		return body(runContext(g, r), r)
	})
}

// RunThenRule is RunThen for a run whose legs are permanent picks with a
// set rule over what a seat picks (game.RegisterPromptRunThenRule): the
// rule is a function of the cards alone and captures nothing.
func RunThenRule(key string, rule game.SetRuleFunc, body func(ctx *Context, r game.RunDone) error) game.PromptRunThen {
	return game.RegisterPromptRunThenRule(key, rule, func(g *game.Game, r game.RunDone) error {
		return body(runContext(g, r), r)
	})
}

// runContext is the Context a keyed run's continuation runs against: a
// copy of the asking item when the run carried one, else a stand-in whose
// controller is the first ID on the carry (keyedContext).
func runContext(g *game.Game, r game.RunDone) *Context {
	if item := r.Item(); item != nil {
		return NewContext(g, item)
	}
	var controller uuid.UUID
	if len(r.Carry) > 0 {
		controller = r.Carry[0]
	}
	return keyedContext(g, controller, r.Source)
}

// packIDLists joins lists of IDs into one carry, separated by uuid.Nil (an
// ID is never Nil). unpackIDLists splits it again, always into as many
// lists as went in; a card whose continuation reads several lists carries
// them this way.
func packIDLists(lists ...[]uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for i, l := range lists {
		out = append(out, l...)
		if i < len(lists)-1 {
			out = append(out, uuid.Nil)
		}
	}
	return out
}

// unpackIDLists is packIDLists' inverse; nil for an empty carry.
func unpackIDLists(carry []uuid.UUID) [][]uuid.UUID {
	if len(carry) == 0 {
		return nil
	}
	lists := [][]uuid.UUID{nil}
	for _, id := range carry {
		if id == uuid.Nil {
			lists = append(lists, nil)
			continue
		}
		lists[len(lists)-1] = append(lists[len(lists)-1], id)
	}
	return lists
}

// unpackIDPair is unpackIDLists for a carry made of exactly two lists
// (packIDLists(a, b)): both, nil where one was empty.
func unpackIDPair(carry []uuid.UUID) (first, second []uuid.UUID) {
	lists := unpackIDLists(carry)
	if len(lists) > 0 {
		first = lists[0]
	}
	if len(lists) > 1 {
		second = lists[1]
	}
	return first, second
}

// RunData is the plain data a keyed run hands its continuation: this
// effect's item as data, the IDs and the numbers the body reads.
func (c *Context) RunData(carry []uuid.UUID, nums ...int) game.RunData {
	return game.RunData{Carry: carry, Nums: nums, Asked: game.NewAskedItem(c.Item)}
}

// PlayersSacrificeThen is "each of <players> sacrifices a permanent of
// their choice, then <then>" with the rest of the card keyed (ADR 0175's
// ordered collector, game.RunData, see RunThen): the seats are asked one at
// a time in APNAP order, each seeing the earlier choices, and everything
// chosen is sacrificed together. A name listed N times owes N permanents.
// `data` is Context.RunData.
func (c *Context) PlayersSacrificeThen(players []uuid.UUID, spec *game.TargetSpec, reason string, then game.PromptRunThen, data game.RunData) error {
	return c.Game.PlayersSacrificeThenKeyedForEffect(c.Source(), players, spec, reason, then, data)
}

// SacrificeThen is "<player> sacrifices <count> permanents of their
// choice, then <then>" with the rest of the card keyed (game.RunData, see
// RunThen): the sacrifice run for one seat.
func (c *Context) SacrificeThen(player uuid.UUID, spec *game.TargetSpec, reason string, count int, then game.PromptRunThen, carry ...uuid.UUID) error {
	return c.Game.PlayerSacrificesThenKeyedDataForEffect(c.Source(), player, spec, reason, count, then, c.RunData(carry))
}
