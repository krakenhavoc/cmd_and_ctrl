package effects

import (
	"fmt"
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eventide's Shadow — Sorcery {4}{B}, (Lorwyn Eclipsed Commander, #5368):
//
//	"Remove any number of counters from among permanents on the battlefield.
//	 You draw cards and lose life equal to the number of counters removed this
//	 way."
//
// As the spell resolves the permanents with counters, and the kinds each has,
// are listed once (every battlefield permanent, anyone's, in battlefield order,
// the kinds sorted); the caster is then asked, for each kind on each permanent
// still there, how many of those counters to remove, 0 up to as many as it has
// as the question is asked (Rhys, the Evermore's chain of option picks). The
// list and the running total are a step record, so a table restored in the
// middle of the questions finds them, and the questions are keyed. With none
// left to ask, the caster draws that many cards and loses that much life, in
// that order, the number being the counters actually removed.
//
// No simplification.
func init() {
	eventidesShadowRemoved = OptionPickThen("option-pick/eventides-shadow-remove", eventidesShadowRemovedBody)
	Register(Spec{
		OracleID:     "92ac8aa4-e23c-4ed9-8317-c249a17ce3d2",
		Name:         "Eventide's Shadow",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			var kinds []string
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				names := make([]string, 0, len(c.Counters))
				for kind, n := range c.Counters {
					if n > 0 {
						names = append(names, kind)
					}
				}
				sort.Strings(names)
				for _, kind := range names {
					ids = append(ids, c.InstanceID)
					kinds = append(kinds, kind)
				}
			}
			ctx.RecordStepRecord(eventidesShadowRecord, game.StepRecord{IDs: ids, Strs: kinds, Ns: []int{0, 0}})
			return eventidesShadowAskNext(ctx, item.ID)
		},
	})
}

// eventidesShadowRecord files the (permanent, kind) pairs, parallel in IDs and
// Strs, with Ns = [how many pairs have been asked about, counters removed so
// far].
const eventidesShadowRecord = "eventides-shadow/pairs"

// eventidesShadowAskNext asks about the next pair whose permanent is still on
// the battlefield with counters of that kind, and files that it is asked. With
// none left it draws and loses life for the total removed.
func eventidesShadowAskNext(ctx *Context, itemID uuid.UUID) error {
	g := ctx.Game
	holder := &game.StackItem{ID: itemID}
	rec, ok := g.StepRecordForEffect(holder, eventidesShadowRecord)
	if !ok || len(rec.Ns) < 2 {
		return nil
	}
	next := rec.Ns[0]
	for ; next < len(rec.IDs) && next < len(rec.Strs); next++ {
		c, ok := g.LookupCardForEffect(rec.IDs[next])
		if z := g.FindCardZoneForEffect(rec.IDs[next]); !ok || z == nil || z.Kind != game.ZoneBattlefield {
			continue
		}
		kind := rec.Strs[next]
		n := c.Counters[kind]
		if n <= 0 {
			continue
		}
		rec.Ns[0] = next + 1
		g.RecordStepForEffect(holder, eventidesShadowRecord, rec)
		opts := make([]game.ChoiceOption, n+1)
		for i := range opts {
			opts[i] = game.ChoiceOption{Label: fmt.Sprintf("Remove %d", i)}
		}
		return PickOption{
			Question: fmt.Sprintf("Eventide's Shadow — remove how many %s counters from %s?", kind, c.Name),
			Options:  opts,
			ThenKey:  eventidesShadowRemoved,
			Carry:    []uuid.UUID{itemID},
		}.Apply(ctx)
	}
	if removed := rec.Ns[1]; removed > 0 {
		if err := (DrawCards{Player: ctx.Controller(), N: removed}).Apply(ctx); err != nil {
			return err
		}
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -removed)
	}
	return nil
}

// eventidesShadowRemoved takes the counters the answer names off the pair just
// asked about, then asks the next. It is registered in init: it asks the next
// question, so it refers to itself.
var eventidesShadowRemoved game.OptionPickThen

func eventidesShadowRemovedBody(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 {
		return nil
	}
	itemID := r.Carry[0]
	holder := &game.StackItem{ID: itemID}
	rec, ok := ctx.Game.StepRecordForEffect(holder, eventidesShadowRecord)
	if !ok || len(rec.Ns) < 2 || rec.Ns[0] < 1 || rec.Ns[0] > len(rec.IDs) {
		return nil
	}
	if r.Index > 0 {
		pair := rec.Ns[0] - 1
		before := counterCountOn(ctx.Game, rec.IDs[pair], rec.Strs[pair])
		if err := (AddCounter{Target: rec.IDs[pair], Kind: rec.Strs[pair], N: -r.Index}).Apply(ctx); err != nil {
			return err
		}
		// What was removed is what actually came off: a replacement may
		// have changed the count.
		rec.Ns[1] += before - counterCountOn(ctx.Game, rec.IDs[pair], rec.Strs[pair])
		ctx.Game.RecordStepForEffect(holder, eventidesShadowRecord, rec)
	}
	return eventidesShadowAskNext(ctx, itemID)
}
