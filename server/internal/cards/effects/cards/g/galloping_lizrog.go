package cardsg

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Galloping Lizrog — Creature — Frog Lizard {3}{G}{U}, 3/3 (Ravnica
// Allegiance, #5521):
//
//	"Trample
//	 When this creature enters, you may remove any number of +1/+1 counters
//	 from among creatures you control. If you do, put twice that many +1/+1
//	 counters on this creature."
//
// The player is asked how many to take off each creature that has any
// (naming none anywhere declines the "may"); they come off as they are
// named, and the last answer puts twice the total on the Lizrog if it is
// still the same object on the battlefield.
//
// No simplification.
func init() {
	const name = "Galloping Lizrog"
	effects.Register(effects.Spec{
		OracleID:        "3561cd52-961f-4098-9287-58ee43970f0f",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(name+" — you may remove any number of +1/+1 counters from among creatures you control; if you do, put twice that many on this creature",
				func(g *game.Game, item *game.StackItem) error {
					return removePlusOneCountersThenGrowSource(effects.NewContext(g, item), name)
				}),
		},
	})
}

// removePlusOneCountersThenGrowSource is "you may remove any
// number of +1/+1 counters from among creatures you control. If you do, put
// twice that many +1/+1 counters on <the source>" (Galloping Lizrog). The
// controller is asked, creature by creature, how many to take off it; each
// answer takes them at once, and the last answer puts twice the total on the
// source. Taking none from every creature is the "may" declined.
func removePlusOneCountersThenGrowSource(ctx *effects.Context, name string) error {
	if effects.SourceIsNewObject(ctx.Game, ctx.Item) || !effects.OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	var queue []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.IsCreature() && c.Counters[game.CounterPlusOne] > 0 {
			queue = append(queue, c.InstanceID)
		}
	}
	return lizrogAskNext(ctx, name, ctx.Source(), queue, 0)
}

// lizrogAskNext asks about the next creature in the queue that
// still has counters, or finishes when none is left. The card name rides the
// prompt as plain data, so the table stays a restore point while it waits.
func lizrogAskNext(ctx *effects.Context, name string, source uuid.UUID, queue []uuid.UUID, total int) error {
	for len(queue) > 0 {
		c, ok := ctx.Game.LookupCardForEffect(queue[0])
		if !ok || c.Counters[game.CounterPlusOne] <= 0 || !effects.OnBattlefield(ctx.Game, queue[0]) {
			queue = queue[1:]
			continue
		}
		options := make([]game.ChoiceOption, 0, c.Counters[game.CounterPlusOne]+1)
		options = append(options, game.ChoiceOption{Label: "Remove none"})
		for i := 1; i <= c.Counters[game.CounterPlusOne]; i++ {
			options = append(options, game.ChoiceOption{Label: "Remove " + strconv.Itoa(i)})
		}
		return effects.PickOption{
			Question: name + " — remove how many +1/+1 counters from " + c.Name + "?",
			Options:  options,
			ThenKey:  lizrogHowMany,
			Carry:    append([]uuid.UUID{source}, queue...),
			Words:    []string{name},
			Nums:     []int{total},
		}.Apply(ctx)
	}
	return lizrogGrow(ctx, source, total)
}

// lizrogHowMany is registered in init: its body asks the next question,
// which names it, so a package-level initializer would be a cycle.
var lizrogHowMany game.OptionPickThen

func init() {
	lizrogHowMany = effects.OptionPickThen("option-pick/galloping-lizrog-how-many", lizrogHowManyBody)
}

func lizrogHowManyBody(ctx *effects.Context, r game.OptionPicked) error {
	if r.Index < 0 || len(r.Carry) < 2 || len(r.Words) == 0 || len(r.Nums) == 0 {
		return nil
	}
	source, creature, rest, total := r.Carry[0], r.Carry[1], r.Carry[2:], r.Nums[0]
	if r.Index > 0 {
		if c, ok := ctx.Game.LookupCardForEffect(creature); ok && c.Counters[game.CounterPlusOne] >= r.Index && effects.OnBattlefield(ctx.Game, creature) {
			if err := ctx.Game.AddCounterForEffect(creature, game.CounterPlusOne, -r.Index); err != nil {
				return err
			}
			total += r.Index
		}
	}
	return lizrogAskNext(ctx, r.Words[0], source, rest, total)
}

// lizrogGrow puts twice `total` +1/+1 counters on the source, if
// it is still the same object on the battlefield.
func lizrogGrow(ctx *effects.Context, source uuid.UUID, total int) error {
	if total <= 0 || !effects.OnBattlefield(ctx.Game, source) {
		return nil
	}
	return effects.AddCounter{Target: source, Kind: game.CounterPlusOne, N: 2 * total}.Apply(ctx)
}
