package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Selesnya Loft Gardens — Plane — Ravnica (Commander Masters, #4870):
//
//	"If an effect would create one or more tokens, it creates twice that
//	 many of those tokens instead.
//	 If an effect would put one or more counters on a permanent, it puts
//	 twice that many of those counters on that permanent instead.
//	 Whenever chaos ensues, until end of turn, whenever you tap a land for
//	 mana, add one mana of any type that land produced."
//
// A plane (ADR 0216 §6). The two statics are replacement effects from the
// command zone (CR 901.7) over every player's tokens and every
// permanent's counters (Doubling Season's clauses with "you control"
// dropped): they apply to effects only, so a loyalty cost, combat damage's
// wither and infect counters, a Saga's lore counter and a paid cost are not
// doubled (the rulings), while a permanent entering with counters is. The
// chaos ability (CR 311.7) sets up a triggered mana ability (CR 605.1b)
// for the turn, for the planar controller's lands, and two resolutions are
// two of them, so it is cumulative (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0a4bf88d-106a-413b-9d80-8775c1ff5ded",
		Name:         "Selesnya Loft Gardens",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			AnyPlayersTokensDoubled("Selesnya Loft Gardens: double tokens"),
			AnyPermanentCountersDoubled("Selesnya Loft Gardens: double counters"),
		},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("Selesnya Loft Gardens — until end of turn, whenever you tap a land for mana, add one mana of any type that land produced",
				Do(TurnManaTriggerAnyLand{Label: "Selesnya Loft Gardens — add one mana of any type that land produced"})),
		},
	})
}
