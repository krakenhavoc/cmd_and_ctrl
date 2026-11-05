package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reconnaissance Partition — Artifact Creature — Borg {4}, 3/3 (Star Trek
// Commander, #4974):
//
//	"This creature enters with two charge counters on it.
//	 {T}, Remove a counter from this creature: Choose one —
//	 • Search your library for a basic land card, put it onto the
//	   battlefield tapped, then shuffle.
//	 • Draw a card."
//
// Loch Mare's shapes: the entry counters are a self-entry replacement
// (CR 614.1c), and "a counter" is one counter of any kind the activator
// names (RemoveCountersFromThis("", 1)), so a +1/+1 counter put on it
// later pays as well as a charge counter. The mode is announced as the
// ability is activated (CR 700.2), Aetheric Amplifier's modal row. The
// {T} is a creature's, so summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c52e4704-6edb-477f-b80c-ea8146d4385f",
		Name:         "Reconnaissance Partition",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterCharge, 2, "Reconnaissance Partition: enters with two charge counters"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Remove a counter from this creature: Choose one — search for a basic land and put it onto the battlefield tapped; or draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 1, Draws: 1},
			Cost:    Plus(TapCost(), RemoveCountersFromThis("", 1)),
			Modes: ChooseOne(
				ModeDoing("Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.", nil,
					func(item *game.StackItem, ctx *Context, _ int) error {
						return fetchBasicTapped(ctx.Game, item)
					}),
				ModeDoing("Draw a card.", nil,
					func(_ *game.StackItem, ctx *Context, _ int) error {
						return DrawCards{N: 1}.Apply(ctx)
					}),
			),
		}},
	})
}
