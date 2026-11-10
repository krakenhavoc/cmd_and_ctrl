package effects

import (
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Slumbering Walker — {3}{W}{W} Creature — Giant Warrior 4/7 (Lorwyn
// Eclipsed, #4458):
//
//	"This creature enters with two -1/-1 counters on it.
//	 At the beginning of your end step, you may remove a counter from this
//	 creature. When you do, return target creature card with power 2 or
//	 less from your graveyard to the battlefield."
//
// The entry counters are Glen Elendra Guardian's replacement. The end-step
// trigger's "you may" is its own (Optional), and "a counter" is one counter
// of a kind you pick (Leatherhead, Swamp Stalker's shape): with one kind on
// the Walker it comes off without a question, with several you choose
// which. Only once a counter has really come off is the CR 603.12
// reflexive trigger created, and its target is chosen as it goes on the
// stack (the ruling), so a Walker with no counters left returns nothing.
// The power is the card's in the graveyard, a `*` card read through its
// own characteristic-defining ability (PowerAnywhereForEffect).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6d972501-4174-4bf0-a26b-2edef31b3a0c",
		Name:         "Slumbering Walker",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterMinusOne, 2, "Slumbering Walker: enters with two -1/-1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			Optional(
				AtYourEndStep("Slumbering Walker — you may remove a counter from this creature",
					func(g *game.Game, item *game.StackItem) error {
						if !sourceIsStillThisPermanent(g, item) {
							return nil
						}
						return slumberingWalkerRemoveACounter(NewContext(g, item), item.SourceCardID)
					}),
				"Slumbering Walker — remove a counter from it to return a creature card with power 2 or less?"),
		},
	})
}

// slumberingWalkerRemoveACounter removes one counter of a kind you pick
// from the Walker, then creates the reflexive return.
func slumberingWalkerRemoveACounter(ctx *Context, self uuid.UUID) error {
	c, ok := ctx.Game.LookupCardForEffect(self)
	if !ok {
		return nil
	}
	var kinds []string
	for kind, n := range c.Counters {
		if n > 0 {
			kinds = append(kinds, kind)
		}
	}
	sort.Strings(kinds)
	switch len(kinds) {
	case 0:
		return nil
	case 1:
		return slumberingWalkerRemoveThenReturn(ctx, self, kinds[0])
	}
	options := make([]game.ChoiceOption, len(kinds))
	for i, kind := range kinds {
		options[i] = game.ChoiceOption{Label: removeACounterPrefix + kind + removeACounterSuffix}
	}
	return PickOption{
		Question: "Slumbering Walker — remove which counter from it?",
		Options:  options,
		ThenKey:  slumberingWalkerCounterPicked,
		Carry:    []uuid.UUID{self},
		WithItem: true,
	}.Apply(ctx)
}

// slumberingWalkerCounterPicked is the kind chosen. An on-disk key.
var slumberingWalkerCounterPicked = OptionPickThen("option-pick/slumbering-walker-remove-a-counter", func(ctx *Context, r game.OptionPicked) error {
	if r.Option == nil || len(r.Carry) != 1 {
		return nil
	}
	kind, ok := strings.CutPrefix(r.Option.Label, removeACounterPrefix)
	if !ok {
		return nil
	}
	return slumberingWalkerRemoveThenReturn(ctx, r.Carry[0], strings.TrimSuffix(kind, removeACounterSuffix))
})

// slumberingWalkerRemoveThenReturn removes the counter and, when it really
// came off, creates "when you do, return target creature card with power 2
// or less from your graveyard to the battlefield".
func slumberingWalkerRemoveThenReturn(ctx *Context, self uuid.UUID, kind string) error {
	if !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(self)
	if !ok || c.Counters[kind] <= 0 {
		return nil
	}
	before := c.Counters[kind]
	if err := (AddCounter{Target: self, Kind: kind, N: -1}).Apply(ctx); err != nil {
		return err
	}
	if after, ok := ctx.Game.LookupCardForEffect(self); !ok || after.Counters[kind] >= before {
		return nil
	}
	return WhenYouDo("Slumbering Walker — return target creature card with power 2 or less from your graveyard to the battlefield",
		slumberingWalkerReturnBody).Apply(ctx)
}

// slumberingWalkerReturnBody is the reflexive half. An on-disk key.
var slumberingWalkerReturnBody = game.ReflexiveBody("slumbering-walker/return-from-graveyard",
	simpleBody(returnFirstLegalGraveyardTargetToBattlefield),
	constTargets(func() *game.TargetSpec {
		return TargetCardInGraveyard("target creature card with power 2 or less from your graveyard",
			YouOwn(), Creature(), func(g *game.Game, _ uuid.UUID, c game.Card) bool {
				return g.PowerAnywhereForEffect(c) <= 2
			})
	}))
