package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feral Hydra — Creature — Hydra Beast {X}{G}, 0/0:
//
//	"This creature enters with X +1/+1 counters on it.
//	 {3}: Put a +1/+1 counter on this creature. Any player may activate
//	 this ability."
//
// The X counters are the printed CR 614.1c entry clause, seeded from
// the cast (XCounters, Hangarback Walker's shape): they land as the
// Hydra enters, so a counter doubler applies. Cast for X=0 it enters as
// a 0/0 and dies to the next state-based check (CR 704.5f).
//
// The activation is an any-player row (CR 602.2, 602.1b): whoever
// activates it pays the {3} out of their own pool (CR 602.1a) and puts
// the counter on the Hydra, whoever controls it. A Hydra that left the
// battlefield, or left and came back as a new object (CR 400.7), gets
// nothing.
//
// No amounts declared: the bot never pays to grow a creature it does not
// control (ADR 0106 owner decision 2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "b5d5dd1f-2d31-4849-8cb9-08354dd48d31",
		Name:                       "Feral Hydra",
		Completeness:               CompletenessFull,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		XMatters:                   true,
		Activated: []ActivatedAbility{{
			Label:     "{3}: Put a +1/+1 counter on this creature. Any player may activate this ability.",
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      ManaCost("{3}"),
			AnyPlayer: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
