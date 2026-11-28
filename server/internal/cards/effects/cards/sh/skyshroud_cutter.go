package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skyshroud Cutter — Creature — Beast {3}{G}, 2/2 (Nemesis, S82 OLD batch
// 680, #5618):
//
//	"If you control a Forest, rather than pay this spell's mana cost, you may
//	 have each other player gain 5 life."
//
// The alternative cost of #6214 (CR 118.9): each player other than the
// caster gains 5 life as the spell is cast, the caster none, offered only
// while the caster controls a Forest (a land with the subtype counts).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fbd71691-d0fd-4842-913b-fe6096d509a3",
		Name:         "Skyshroud Cutter",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{{
			Key:                      "skyshroud-cutter",
			Label:                    "If you control a Forest, have each other player gain 5 life rather than pay this spell's mana cost",
			EachOtherPlayerGainsLife: 5,
			Condition:                effects.OldB680YouControlAForest,
		}},
	})
}
