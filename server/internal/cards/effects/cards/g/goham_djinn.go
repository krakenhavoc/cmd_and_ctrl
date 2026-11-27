package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goham Djinn — Creature — Djinn {5}{B}, 5/5 (Invasion, S82 OLD #5616):
//
//	"{1}{B}: Regenerate this creature.
//	 This creature gets -2/-2 as long as black is the most common color among
//	 all permanents or is tied for most common."
//
// The regeneration is the shared regenerate row (a shield, CR 701.19a). The
// -2/-2 is the Mirage Djinns' layer 7c static
// (effects.SelfPumpWhileColorIsMostCommon): it counts the layered colours of
// every permanent on the battlefield, this creature's included (CR 105.2), and
// holds while black is first or tied for first, so it comes and goes as the
// board changes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9db02fb-78e2-44b3-9e1b-ef0148234337",
		Name:         "Goham Djinn",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.RegenerateThisRow("{1}{B}")},
		Static:       []game.StaticAbility{effects.SelfPumpWhileColorIsMostCommon("B", -2, -2)},
	})
}
