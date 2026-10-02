package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Floodchaser — Creature — Elemental {5}{U}, 0/0:
//
//	"This creature enters with six +1/+1 counters on it.
//	 This creature can't attack unless defending player controls an
//	 Island.
//	 {U}, Remove a +1/+1 counter from this creature: Target land becomes
//	 an Island until end of turn."
//
// The counters are a CR 614.1c entry replacement. The restriction is ADR
// 0107 §2's (#1879, CR 508.1c), read per defending player (CR 508.5). The
// ability is ADR 0109 §1's (#1881) CR 305.7 type set: until end of turn
// the land is an Island (its other subtypes stay, CR 205.1a), loses its
// rules-text abilities and taps for {U} (CR 305.6). Each activation costs
// it a counter, so it shrinks as it opens the way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1bf97485-191b-466c-9590-3449466b348c",
		Name:         "Floodchaser",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{b10EntersWithCounters(game.CounterPlusOne, 6, "Floodchaser: enters with six +1/+1 counters")},
		Static: []game.StaticAbility{
			CantAttackUnlessDefendingPlayerControls(QuerySubtype("Island")),
		},
		Activated: []ActivatedAbility{{
			Label:   "{U}, Remove a +1/+1 counter from this creature: Target land becomes an Island until end of turn.",
			Cost:    Plus(ManaCost("{U}"), RemoveCountersFromThis(game.CounterPlusOne, 1)),
			Targets: TargetPermanent("target land", Land()),
			Effect:  TargetLandBecomesUntilEOT("Floodchaser", "Island"),
		}},
	})
}
