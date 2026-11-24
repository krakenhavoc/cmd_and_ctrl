package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sojourner's Companion — Artifact Creature — Salamander {7}, 4/4 (Modern
// Horizons 2, #5501):
//
//	"Affinity for artifacts
//	 Artifact landcycling {2}"
//
// Affinity is the self cost reduction of CR 702.41a over artifacts you
// control. Artifact landcycling is typecycling (CR 702.29e): the search is
// for an artifact land card, which must be both.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1cbfe1f2-776b-429d-9a60-8da32058a927",
		Name:         "Sojourner's Companion",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for artifacts", effects.Artifact()),
		},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Artifact landcycling", "{2}", "an artifact land card",
				func(c game.Card) bool { return c.IsArtifact() && c.IsLand() }),
		},
	})
}
