package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Herald of Kozilek — Creature — Eldrazi Drone {1}{U}{R}, 2/4 (Battle for
// Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 Colorless spells you cast cost {1} less to cast."
//
// Eye of Ugin's colorless-spell test on the spell being cast, as a one-generic
// reduction that applies to your spells only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e1890cbc-b15b-44bc-b1c0-ca3d9fd6028a",
		Name:            "Herald of Kozilek",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Colorless spells you cast cost {1} less to cast.",
				effects.YourSpell(), func(q game.CostQuery) bool { return q.Card.IsColorless() }),
		},
	})
}
