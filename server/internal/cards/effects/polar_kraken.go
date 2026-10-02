package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Polar Kraken — Creature — Kraken, {8}{U}{U}{U}, 11/11:
//
//	"Trample
//	 This creature enters tapped.
//	 Cumulative upkeep—Sacrifice a land. (At the beginning of your upkeep, put an age counter on this permanent, then sacrifice it unless you pay its upkeep cost for each age counter on it.)"
//
// Cumulative upkeep with a sacrifice (CR 702.24a, ADR 0108 §5): with N age
// counters the controller sacrifices N lands, chosen and paid together, or
// sacrifices the Kraken. "Enters tapped" is SelfEntersTapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d0ff30c8-ddb7-439d-b14d-c83fe3c8da87",
		Name:            "Polar Kraken",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements:    []game.ReplacementEffect{SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			CumulativeUpkeepPaying("Polar Kraken — cumulative upkeep: sacrifice a land",
				SacrificePayment(1, "land", "lands", game.PermanentQuery{Types: []string{"land"}}), "land", "lands"),
		},
	})
}
