package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cat-Owl — Creature — Cat Bird {3}{W/U}, 3/3 (Avatar: The Last
// Airbender, #4459):
//
//	"Flying
//	 Whenever this creature attacks, untap target artifact or creature."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d); it
// may be the Cat-Owl itself, which attacking tapped. A target gone or no
// longer an artifact or creature as it resolves is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ba6840a6-b6fd-4b91-ac1a-a61b7b15849a",
		Name:            "Cat-Owl",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Cat-Owl — untap target artifact or creature", effects.UntapTheTarget),
				effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature()))),
		},
	})
}
