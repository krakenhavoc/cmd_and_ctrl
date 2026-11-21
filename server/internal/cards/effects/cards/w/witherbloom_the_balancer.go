package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Witherbloom, the Balancer — Legendary Creature — Elder Dragon
// {6}{B}{G}, 5/5:
//
//	"Affinity for creatures (This spell costs {1} less to cast for
//	 each creature you control.)
//	 Flying, deathtouch
//	 Instant and sorcery spells you cast have affinity for creatures."
//
// #746: both slots. Its own affinity is a self cost modifier; the
// grant is an ordinary battlefield cost modifier (ADR 0048 §1), one
// instance per Witherbloom, so two of them both apply (CR 702.41b).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fbb04a21-e513-4317-aac8-fa6df91c3438",
		Name:            "Witherbloom, the Balancer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for creatures", effects.Creature()),
		},
		CostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.PermanentsYouControl(effects.Creature()),
				"Instant and sorcery spells you cast have affinity for creatures.",
				effects.YourSpell(), effects.InstantOrSorcerySpell()),
		},
	})
}
