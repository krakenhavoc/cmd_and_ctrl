package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hollowhenge Beast — Creature — Beast {3}{G}{G}, 5/5 (Dark Ascension, #5571):
//
//	(no rules text)
//
// Vanilla, as Memnite: registered so the card is audited and published
// as complete. The body, cost and type line are printed card data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a13b02c2-7ec1-4032-99cc-c76fe512914e",
		Name:         "Hollowhenge Beast",
		Completeness: effects.CompletenessFull,
	})
}
