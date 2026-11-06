package roadmap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// declaresCollectEvidence is the collect evidence probe (#5193): a spell
// whose optional additional cost collects evidence
// (effects.OptionalCollectEvidence), or an activated ability whose cost does
// (effects.CollectEvidenceN), or a spell whose mandatory additional cost
// collects evidence sized by its targets (effects.CollectEvidenceFromTargetsCost,
// #5240). The effect form is a registered continuation and has no
// declaration to read; its cards are matched by Printed.
func declaresCollectEvidence(s effects.Spec) bool {
	if s.AdditionalCost != nil && s.AdditionalCost.ExileCards.IsCollectEvidenceFromTargets() {
		return true
	}
	for _, oc := range s.OptionalCosts {
		if oc.Key == game.CollectEvidenceKey && oc.ExileCards.IsCollectEvidence() {
			return true
		}
	}
	for _, a := range s.Activated {
		if a.Cost.ExileCards.IsCollectEvidence() {
			return true
		}
	}
	return false
}
