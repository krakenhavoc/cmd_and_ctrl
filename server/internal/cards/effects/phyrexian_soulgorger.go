package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Soulgorger — Snow Artifact Creature — Phyrexian Construct, {3}, 8/8:
//
//	"Cumulative upkeep—Sacrifice a creature. (At the beginning of your upkeep, put an age counter on this permanent, then sacrifice it unless you pay its upkeep cost for each age counter on it.)"
//
// Cumulative upkeep with a sacrifice (CR 702.24a, ADR 0108 §5): with N age
// counters the controller sacrifices N creatures, chosen and paid together,
// or sacrifices the Soulgorger. It is a creature itself, so it may be one of
// them, as the printed cost allows.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b6e36e77-0cea-4b7a-b968-774b1e74a992",
		Name:         "Phyrexian Soulgorger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			CumulativeUpkeepPaying("Phyrexian Soulgorger — cumulative upkeep: sacrifice a creature",
				SacrificePayment(1, "creature", "creatures", game.PermanentQuery{Types: []string{"creature"}}), "creature", "creatures"),
		},
	})
}
