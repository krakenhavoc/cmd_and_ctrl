package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fractal Tender — Creature — Elf Wizard {3}{G}{U}, 3/3 (Secrets of
// Strixhaven, #4456):
//
//	"Ward {2}
//	 Increment (Whenever you cast a spell, if the amount of mana you spent is
//	 greater than this creature's power or toughness, put a +1/+1 counter on
//	 this creature.)
//	 At the beginning of each end step, if you put a counter on this creature
//	 this turn, create a 0/0 green and blue Fractal creature token and put
//	 three +1/+1 counters on it."
//
// Ward {2} is Ward(WardMana("{2}")); Increment is the keyword helper (CR
// 702.191, Cuboid Colony's shape).
//
// The end-step ability triggers at the beginning of EVERY end step, yours or
// an opponent's, and "if you put a counter on this creature this turn" is an
// intervening if (CR 603.4): asked as the step begins and again as the
// ability resolves, read off this turn's event log
// (socB25YouPutACounterOnThisObjectThisTurn): a counter of any kind that YOU
// put on this very object, so counters another player put, or put on an
// earlier object of the same card, do not count. It triggers once however
// many counters. The Fractal is Deekah's create-then-put shape
// (socB8FractalWithCounters) with three counters.
//
// No simplification.
func init() {
	const label = "Fractal Tender — create a 0/0 Fractal and put three +1/+1 counters on it"
	Register(Spec{
		OracleID:        "cb45860a-1937-4803-b6d8-331a35be39fd",
		Name:            "Fractal Tender",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{KeywordIncrement},
		ReadsManaSpent:  game.ReadsYourCasts,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Fractal Tender — ward {2}"),
			Increment(),
			InterveningIf(
				CondBoard{
					Name: "you put a counter on this creature this turn",
					Test: func(ctx *Context) bool {
						return socB25YouPutACounterOnThisObjectThisTurn(ctx.Game, ctx.Controller(), ctx.Item.SourceCardID)
					},
				},
				On(game.EventBeginEndStep,
					func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						return socB25YouPutACounterOnThisObjectThisTurn(g, source.Controller, source.InstanceID)
					},
					label,
					socB8FractalWithCounters("fractal-tender/fractal", func(*Context) int { return 3 }))),
		},
	})
}
