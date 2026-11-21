package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Firebomb — Artifact {1} (Foundations, #4465):
//
//	"Flash
//	 {7}, {T}, Sacrifice this artifact: Destroy target permanent."
//
// Flash rides PrintedKeywords. The destroy row is Unstable Obelisk's: the
// whole cost (seven mana, the tap and the sacrifice) is paid at announce, so
// the Firebomb is gone when the ability resolves, and the target is checked
// again then (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2625fa9c-638c-46d3-966b-1166f440186f",
		Name:            "Goblin Firebomb",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{7}, {T}, Sacrifice this artifact: Destroy target permanent.",
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{7}"},
			Targets: effects.TargetPermanent("target permanent"),
			Effect:  effects.DestroyFirstLegalCardTarget,
		}},
	})
}
