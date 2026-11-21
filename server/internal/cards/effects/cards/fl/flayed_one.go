package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flayed One — Artifact Creature — Necron {2}{B}, 4/1 (Warhammer 40,000
// Commander, S82 GAP batch 187, #5948):
//
//	"Lifelink
//	 Flesh Flayer — When this creature enters, mill three cards."
//
// "Mill" is the controller's own library (CR 701.17a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "899415fd-4164-42a5-a0c7-eef1967c5ec5",
		Name:            "Flayed One",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Flayed One — mill three cards", effects.Do(effects.MillCards{N: 3})),
		},
	})
}
