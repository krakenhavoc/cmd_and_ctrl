package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ingenious Prodigy — Creature — Human Wizard {X}{U}, 0/1:
//
//	"Skulk (This creature can't be blocked by creatures with greater
//	 power.)
//	 This creature enters with X +1/+1 counters on it.
//	 At the beginning of your upkeep, if this creature has one or more
//	 +1/+1 counters on it, you may remove a +1/+1 counter from it. If you
//	 do, draw a card."
//
// Skulk is the engine's block rule. The counters are the X announced for
// the spell (CR 107.3m), so a Prodigy put onto the battlefield without
// being cast has none. The upkeep trigger is an intervening "if"
// (CR 603.4), checked as the upkeep begins and again before the "you
// may" is asked; the draw happens only when a counter really came off
// this same object.
//
// No simplification.
func init() {
	const label = "Ingenious Prodigy — you may remove a +1/+1 counter from it; if you do, draw a card"
	Register(Spec{
		OracleID:                   "6f498ff7-38c9-44ea-9cb5-994a003825ee",
		Name:                       "Ingenious Prodigy",
		Completeness:               CompletenessFull,
		PrintedKeywords:            []string{"skulk"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		XMatters:                   true,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "this creature has one or more +1/+1 counters on it",
				Test: func(ctx *Context) bool { return socB7SourceHasPlusOneCounter(ctx.Game, ctx.Item.SourceCardID) },
			}, Optional(On(game.EventBeginUpkeep, AllOf(ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return socB7SourceHasPlusOneCounter(g, source.InstanceID)
			}), label, Do(Instruction(ingeniousProdigyRemoveAndDraw))),
				"Ingenious Prodigy — remove a +1/+1 counter from it to draw a card?")),
		},
	})
}

// ingeniousProdigyRemoveAndDraw removes the counter from the Prodigy, if
// it is still the object that triggered and still has one, and draws only
// then.
func ingeniousProdigyRemoveAndDraw(ctx *Context) error {
	if !sourceIsStillThisPermanent(ctx.Game, ctx.Item) || !socB7SourceHasPlusOneCounter(ctx.Game, ctx.Item.SourceCardID) {
		return nil
	}
	if err := ctx.Game.AddCounterForEffect(ctx.Item.SourceCardID, game.CounterPlusOne, -1); err != nil {
		return err
	}
	return DrawCards{N: 1}.Apply(ctx)
}
