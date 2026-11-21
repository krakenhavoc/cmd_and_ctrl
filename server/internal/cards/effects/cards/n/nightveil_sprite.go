package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nightveil Sprite — Creature — Faerie Rogue {1}{U}, 1/2:
//
//	"Flying
//	 Whenever this creature attacks, surveil 1."
//
// Flying rides PrintedKeywords; the attack trigger is one surveil.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8ac23863-0aa7-41bc-b9b7-c5f0566172cf",
		Name:            "Nightveil Sprite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Nightveil Sprite — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
	})
}
