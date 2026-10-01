package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chimil, the Inner Sun — Legendary Artifact {6}:
//
//	"Spells you control can't be countered.
//	 At the beginning of your end step, discover 5."
//
// The discover is ADR 0099's (game/discover.go). The first line is
// ADR 0106 §4's battlefield static (#1806): a statement the counter
// gate reads off the battlefield whenever something tries to counter a
// spell, so it covers every spell its controller controls while Chimil
// is out with its abilities — one cast before Chimil arrived, a copy,
// the spell discover casts — and none once Chimil has gone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "119d9671-61ba-4629-89b1-f94bdec5cb74",
		Name:         "Chimil, the Inner Sun",
		Completeness: CompletenessFull,
		Discovers:    true,
		SpellsCantBeCountered: []game.CounterShieldStatic{
			SpellsYouControlCantBeCountered("Spells you control can't be countered."),
		},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Chimil, the Inner Sun — discover 5", DiscoverN(5)),
		},
	})
}
