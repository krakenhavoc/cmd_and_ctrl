package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ambitious Augmenter — Creature — Turtle Wizard {G}, 1/1 (Secrets of
// Strixhaven, #4456):
//
//	"Increment (Whenever you cast a spell, if the amount of mana you
//	 spent is greater than this creature's power or toughness, put a +1/+1
//	 counter on this creature.)
//	 When this creature dies, if it had one or more counters on it, create
//	 a 0/0 green and blue Fractal creature token, then put this creature's
//	 counters on that token."
//
// Increment is the keyword helper (CR 702.191, Increment). The death
// trigger's "if" is an intervening if (CR 603.4) asked of the creature as
// it last existed (CR 603.10a): the engine's departure record of its
// counters. Every kind of counter it had moves, not only +1/+1, read at
// resolution from the same record (CR 608.2h) and put on the new Fractal
// after it exists (Deekah's shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d87f0e52-486e-49b4-9cf1-12e238857340",
		Name:            "Ambitious Augmenter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{KeywordIncrement},
		ReadsManaSpent:  game.ReadsYourCasts,
		Triggered: []game.TriggeredAbility{
			Increment(),
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return cardDied(ev, source) && len(g.LastKnownCountersForEffect(ev.CardID)) > 0
			}, "Ambitious Augmenter — create a Fractal and put this creature's counters on it",
				socB19FractalWithTheDeadsCounters("ambitious-augmenter/fractal")),
		},
	})
}
