package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crow of Dark Tidings — Creature — Zombie Bird {2}{B}, 2/1
// (Foundations, #4465):
//
//	"Flying
//	 When this creature enters or dies, mill two cards."
//
// Stitcher's Supplier's one ability with two conditions
// (WhenThisEntersOrDies), milling two. Flying rides PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a4ffe297-5e82-43f9-91a4-7aa3d8dd3b4a",
		Name:            "Crow of Dark Tidings",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrDies("Crow of Dark Tidings — mill two cards", effects.Do(effects.MillCards{N: 2})),
		},
	})
}
