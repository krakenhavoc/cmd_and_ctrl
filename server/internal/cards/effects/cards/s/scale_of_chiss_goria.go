package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scale of Chiss-Goria — Artifact {3} (Mirrodin, #5606):
//
//	"Flash
//	 Affinity for artifacts (This spell costs {1} less to cast for each
//	 artifact you control.)
//	 {T}: Target creature gets +0/+1 until end of turn."
//
// Myr Enforcer's affinity over an artifact with flash, and Mishra's Factory's
// targeted pump as a {T} ability. The artifact is not a creature, so it
// taps the turn it arrives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "1ab0a315-0d0d-4247-9bb9-107b35b353d6",
		Name:              "Scale of Chiss-Goria",
		Completeness:      effects.CompletenessFull,
		PrintedKeywords:   []string{"flash"},
		SelfCostModifiers: []game.CostModifier{effects.AffinityFor("Affinity for artifacts", effects.Artifact())},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature gets +0/+1 until end of turn.",
			Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 0, 1)), game.AnswerPump),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.BoostTargetUntilEOT("Scale of Chiss-Goria — +0/+1", 0, 1),
		}},
	})
}
