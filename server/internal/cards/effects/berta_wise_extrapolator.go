package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Berta, Wise Extrapolator — Legendary Creature — Frog Druid {2}{G}{U},
// 1/4 (Secrets of Strixhaven, #4456):
//
//	"Increment (Whenever you cast a spell, if the amount of mana you spent
//	 is greater than this creature's power or toughness, put a +1/+1
//	 counter on this creature.)
//	 Whenever one or more +1/+1 counters are put on Berta, add one mana of
//	 any color.
//	 {X}, {T}: Create a 0/0 green and blue Fractal creature token and put
//	 X +1/+1 counters on it."
//
// Increment is the keyword helper (CR 702.191, Increment). "One or more
// ... are put" is one trigger per placement event whose +1/+1 delta is
// above zero (Scurry Oak's test), so an Increment counter, a counter from
// another source and a counter put on Berta as she enters each add one
// mana; a placement of some other kind of counter does not. The mana is
// added as the ability resolves, not as a mana ability (the ruling), so it
// uses the stack and empties with the step (CR 106.4). The activated
// ability creates the Fractal and then puts X counters on it (Deekah's
// shape, socB8FractalWithCounters); with X = 0 the token is a 0/0 and dies
// to state-based actions (CR 704.5f). It is not a mana ability either.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "715f0832-696c-4111-b418-d9b6da403159",
		Name:            "Berta, Wise Extrapolator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{KeywordIncrement},
		ReadsManaSpent:  game.ReadsYourCasts,
		XMatters:        true,
		Triggered: []game.TriggeredAbility{
			Increment(),
			On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Target == source.InstanceID && b33CountersPlacedDelta(ev, game.CounterPlusOne, g) > 0
			}, "Berta, Wise Extrapolator — add one mana of any color",
				Do(AddMana{Produced: "{W|U|B|R|G}"})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{X}, {T}: Create a 0/0 green and blue Fractal creature token and put X +1/+1 counters on it.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1, CreatureTokens: 1},
			Cost:    Plus(ManaCost("{X}"), TapCost()),
			Effect: socB8FractalWithCounters("berta-wise-extrapolator/fractal", func(ctx *Context) int {
				return ctx.X()
			}),
		}},
	})
}
