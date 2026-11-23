package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Illusionary Servant — Creature — Illusion {1}{U}{U}, 3/4 (Magic 2010, #5585):
//
//	"Flying
//	 When this creature becomes the target of a spell or ability,
//	 sacrifice it."
//
// Flying rides PrintedKeywords. The sacrifice is Phantasmal Shieldback's
// trigger: it fires when the Servant becomes the target of any spell or
// ability, and resolves before that spell or ability does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "28e99c5f-a78a-4323-b5e9-1b43303acaf6",
		Name:            "Illusionary Servant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, effects.Self,
				"Illusionary Servant — sacrifice it",
				effects.SacrificeThisIfStillOnBattlefield),
		},
	})
}
