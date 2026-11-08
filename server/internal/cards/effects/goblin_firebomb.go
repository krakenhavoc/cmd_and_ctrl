package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "2625fa9c-638c-46d3-966b-1166f440186f",
		Name:            "Goblin Firebomb",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Activated: []ActivatedAbility{{
			Label:   "{7}, {T}, Sacrifice this artifact: Destroy target permanent.",
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{7}"},
			Targets: TargetPermanent("target permanent"),
			Effect:  destroyFirstLegalCardTarget,
		}},
	})
}
