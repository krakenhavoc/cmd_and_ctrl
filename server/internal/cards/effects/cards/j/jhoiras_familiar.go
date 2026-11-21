package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jhoira's Familiar — Artifact Creature — Bird {4}, 2/2:
//
//	"Flying
//	 Historic spells you cast cost {1} less to cast. (Artifacts,
//	 legendaries, and Sagas are historic.)"
//
// Flying rides PrintedKeywords. "Historic" (CR 301.2 / the parenthetical
// itself) has no shared predicate yet, so it's a small closure local to
// this card: artifact, legendary, or a Saga (HasSubtype, effective
// subtypes).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "df4b5f7b-9d83-49b1-bd5e-77d0652eb34c",
		Name:            "Jhoira's Familiar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Historic spells you cast cost {1} less to cast.",
				effects.YourSpell(), historicSpell()),
		},
	})
}

// historicSpell is "artifacts, legendaries, and Sagas are historic" —
// the reminder text on Jhoira's Familiar and every other historic
// payoff.
func historicSpell() effects.CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Card.IsArtifact() || q.Card.IsLegendary() || q.Card.HasSubtype("Saga")
	}
}
