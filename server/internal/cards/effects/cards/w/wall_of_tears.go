package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wall of Tears — Creature — Wall {1}{U}, 0/4 (Stronghold, S82 OLD b749, #5627):
//
//	"Defender
//	 Whenever this creature blocks a creature, return that creature to its
//	 owner's hand at end of combat."
//
// Kaijin of the Vanishing Touch's block trigger (OldB749WheneverThisBlocks
// ReturnItAtEndOfCombat): once per attacker blocked, returning it only if it is
// still that object and still on the battlefield at the end of combat step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5825d718-6a32-4f4c-9b56-1a50051cfe28",
		Name:            "Wall of Tears",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered:       []game.TriggeredAbility{effects.OldB749WheneverThisBlocksReturnItAtEndOfCombat("Wall of Tears")},
	})
}
