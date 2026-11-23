package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vrock — {3}{B}{B} Creature — Bird Demon 3/3
//
//	"Flying
//	 Toxic Spores — At the beginning of your end step, if a permanent you
//	 controlled left the battlefield this turn, each opponent loses 3 life."
//
// Revolt's condition (CR 207.2c), an intervening "if" read as the end step
// begins and again as the trigger resolves (CR 603.4).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f50d2a5c-2408-4c09-81cd-55e2ca9b1ce1",
		Name:            "Vrock",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStepIfRevolt("Toxic Spores — each opponent loses 3 life", effects.Do(effects.EachOpponentLosesLife{N: 3})),
		},
	})
}
