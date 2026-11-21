package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sneering Shadewriter — Creature — Vampire Warlock {4}{B}, 3/3 (Secrets of
// Strixhaven, #4456):
//
//	"Flying
//	 When this creature enters, each opponent loses 2 life and you gain 2
//	 life."
//
// Flying is a printed keyword; the enters trigger is the drain
// (EachOpponentLosesLife, then GainLife), a life change with no damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "21a094d6-de3d-48d4-a35d-b99617da55db",
		Name:            "Sneering Shadewriter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sneering Shadewriter — each opponent loses 2 life and you gain 2 life",
				effects.Do(effects.EachOpponentLosesLife{N: 2}, effects.GainLife{Amount: 2})),
		},
	})
}
