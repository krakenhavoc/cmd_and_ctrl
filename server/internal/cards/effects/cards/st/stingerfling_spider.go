package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stingerfling Spider — Creature — Spider {4}{G}, 2/5 (Magic 2012, #5573):
//
//	"Reach
//	 When this creature enters, you may destroy target creature with
//	 flying."
//
// Reach, and Hammerheim Deadeye's enters trigger made optional (Optional):
// the target is chosen as it goes on the stack and the destroy is asked
// for as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "056770e2-8ab7-424d-a2d8-7be6d6cc73f1",
		Name:            "Stingerfling Spider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisEnters("Stingerfling Spider — destroy target creature with flying", effects.DestroyFirstLegalCardTarget),
				effects.TargetCreature("target creature with flying", effects.HasKeyword("flying"))),
				"Stingerfling Spider — destroy target creature with flying?"),
		},
	})
}
