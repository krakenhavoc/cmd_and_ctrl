package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heartwood Shard — Artifact {3} (Mirrodin, #5606):
//
//	"{3}, {T} or {G}, {T}: Target creature gains trample until end of
//	 turn."
//
// Crystal Shard's two-row shape (one ability, a choice of cost, picked by
// picking the row). Both rows give the target trample until end of turn
// as a layer-6 grant pinned to it (ebTargetCreatureGainsUntilEOT); a
// target that is no longer legal as the ability resolves gets nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	const effect = "Target creature gains trample until end of turn."
	grant := effects.EbTargetCreatureGainsUntilEOT("trample", "Heartwood Shard — trample until end of turn")
	effects.Register(effects.Spec{
		OracleID:     "64d4bdc5-ff40-469d-bb25-82de0a49dfa4",
		Name:         "Heartwood Shard",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}, {T}: " + effect,
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "trample")),
				Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  grant,
			},
			{
				Label:   "{G}, {T}: " + effect,
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "trample")),
				Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  grant,
			},
		},
	})
}
