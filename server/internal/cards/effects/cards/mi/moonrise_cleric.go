package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moonrise Cleric — Creature — Bat Cleric {1}{W/B}{W/B}, 2/3
// (Bloomburrow, #4467):
//
//	"Flying
//	 Whenever this creature attacks, you gain 1 life."
//
// Flying is the printed keyword; the attack trigger is the shared
// WheneverThisAttacks shape (Teacher's Pest's), which watches this
// creature's own attack declaration only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d8067f42-68c0-42da-a1d6-859953939555",
		Name:            "Moonrise Cleric",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Moonrise Cleric — you gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
