package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Exocrine — Creature — Tyranid {X}{2}{R}, 2/2 (Warhammer 40,000 Commander,
// S82 GAP batch 194, #5948):
//
//	"Ravenous (This creature enters with X +1/+1 counters on it. If X is 5 or
//	 more, draw a card when it enters.)
//	 Bio-plasmic Barrage — When this creature enters, it deals X damage to each
//	 player and each other creature."
//
// X is the value announced for the spell (CR 107.3m). The counters are the
// cast-time entry counters (Morlun's XCounters), the ravenous draw is its own
// enters trigger, and the barrage reads X off the permanent as it is built and
// carries it on the item: the damage is one batch to every player still in the
// game and every creature but the Exocrine, dealt by the Exocrine as it last
// existed. A copy or a cheat onto the battlefield has X = 0 and does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "cdf34da6-d952-43ca-855d-fcd9d49efae6",
		Name:                       "Exocrine",
		Completeness:               effects.CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterPlusOne)},
		Purpose:                    game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.R42RavenousDraw("Exocrine"),
			effects.GapB194ExocrineEnters("Exocrine — it deals X damage to each player and each other creature"),
		},
	})
}
