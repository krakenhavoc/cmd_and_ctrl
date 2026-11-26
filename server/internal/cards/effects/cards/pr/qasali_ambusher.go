package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Qasali Ambusher — {1}{G}{W} Creature — Cat Warrior 2/3 (Shards of Alara, S82
// OLD b515):
//
//	"Reach
//	 If a creature is attacking you and you control a Forest and a Plains, you
//	 may cast this spell without paying its mana cost and as though it had
//	 flash."
//
// An alternative cost (CR 118.9) with an empty price, offered only while its
// condition holds; claiming it is what opens the instant-speed window (CR
// 601.3c), so the printed-cost cast stays at sorcery speed. The condition is
// judged at announcement.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cbd106fb-c9c5-4c5e-8801-619b7c8ff5a6",
		Name:            "Qasali Ambusher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		AlternativeCosts: []game.AlternativeCost{{
			Key:           "qasali-ambusher",
			Label:         "If a creature is attacking you and you control a Forest and a Plains, cast this spell without paying its mana cost and as though it had flash",
			AsThoughFlash: true,
			Condition: func(g *game.Game, controller uuid.UUID) bool {
				return effects.OldB515ACreatureIsAttackingYou(g, controller) && effects.OldB515YouControlForestAndPlains(g, controller)
			},
		}},
	})
}
