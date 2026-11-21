package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aven Battle Priest — Creature — Bird Cleric {5}{W}, 3/3 (Magic Origins, #5549):
//
//	"Flying (This creature can't be blocked except by creatures with flying or reach.)
//	 When this creature enters, you gain 3 life."
//
// Angel of Mercy's shape: flying and an entry life gain.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "696fb1fa-5535-4f73-8235-ad4416dcfdd0",
		Name:            "Aven Battle Priest",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Aven Battle Priest — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
	})
}
