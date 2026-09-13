package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pir, Imaginative Rascal — 1/1 Human for {2}{G}:
//
//	"Partner with Toothy, Imaginary Friend
//	 If one or more counters would be put on a permanent your team
//	 controls, that many plus one of each of those kinds of counters
//	 are put on that permanent instead."
//
// The +1 half is Hardened Scales widened from "a creature you
// control" to "a permanent your team controls", which matters: Pir
// adds a loyalty counter to your planeswalker's +1 and a defense
// counter to your battle, and Hardened Scales does neither.
//
// TWO DECLARED SIMPLIFICATIONS, both weaker than printed.
//
// "Your team" is read as "you". Two-Headed Giant is not a format this
// engine has, so a team is one player and the two readings coincide
// at every table it will ever see. Stated rather than silently
// assumed, because the day a team format arrives this is where the
// assumption is.
//
// Partner with is absent. It is a pair of abilities — a
// tutor-on-entry trigger and a Commander deck-construction
// permission — and the engine has neither a "search your library for
// a card by NAME" primitive nor partner-aware commander validation.
// Shipping the tutor half without the deck-construction half would be
// the wrong subset anyway: the trigger is the small half, and the
// reason anybody plays Pir is the +1.
func init() {
	Register(Spec{
		OracleID:     "7683c2b2-a06f-4691-9cc5-1968dc032885",
		Name:         "Pir, Imaginative Rascal",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"\"Partner with Toothy\" is missing — no search trigger, and Pir can't be paired as a commander."},
		Replacements: []game.ReplacementEffect{
			{
				Watches: []game.EventKind{game.EventCounterPlaced},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
					// Placement only, and on any PERMANENT you
					// control — not just a creature, which is the
					// whole difference from Hardened Scales.
					return counterPlacementOn(ev, g, src.Controller, true)
				},
				Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
					ev.CounterDelta++
					return nil
				},
				Controller: vorinclexController,
				Label:      "Pir: one more counter",
			},
		},
	})
}
