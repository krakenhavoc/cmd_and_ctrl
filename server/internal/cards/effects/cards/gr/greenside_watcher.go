package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Greenside Watcher — Creature — Elf Druid {1}{G}, 2/1 (Gatecrash, #5565):
//
//	"{T}: Untap target Gate."
//
// The target is any Gate on the battlefield, whoever controls it, checked
// again as the ability resolves (CR 608.2b). The Watcher taps itself as a
// cost, so it is not itself a Gate to untap.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6297680-5d32-41e6-aaaf-b194e6be6b96",
		Name:         "Greenside Watcher",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Untap target Gate.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target Gate", effects.Subtype("Gate")),
			Effect:  effects.UntapEachLegalTarget,
		}},
	})
}
