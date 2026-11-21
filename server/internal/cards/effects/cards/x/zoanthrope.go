package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zoanthrope — Creature — Tyranid {X}{U}{R}, 0/0 (Warhammer 40,000 Commander,
// S82 GAP batch 195, #5948):
//
//	"Ravenous (This creature enters with X +1/+1 counters on it. If X is 5 or
//	 more, draw a card when it enters.)
//	 Flying, ward {2}
//	 Warp Blast — When this creature enters, it deals X damage to any target."
//
// X is the value announced for the spell (CR 107.3m). The counters are the
// cast-time entry counters (Morlun's XCounters), the ravenous draw is its own
// enters trigger, and Warp Blast reads X off the permanent as it is built and
// carries it on the item, so it is exact even if the Zoanthrope has left; a
// copy or a cheat onto the battlefield has X = 0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "e7d79fbe-b02c-41c7-8b0a-6477feb1da7a",
		Name:                       "Zoanthrope",
		Completeness:               effects.CompletenessFull,
		PrintedKeywords:            []string{"flying"},
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterPlusOne)},
		Triggered: []game.TriggeredAbility{
			effects.R42RavenousDraw("Zoanthrope"),
			effects.Ward(effects.WardMana("{2}"), "Zoanthrope — ward {2}"),
			effects.EtbDealsCastXDamage("Zoanthrope — it deals X damage to any target", effects.TargetAny()),
		},
	})
}
