package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scarblade Scout — Creature — Elf Scout {1}{B}, 2/2 (Lorwyn Eclipsed,
// #4458):
//
//	"Lifelink
//	 When this creature enters, mill two cards."
//
// The enter trigger is the footprinted mill primitive (CR 701.17a): the top
// two cards of your library go to your graveyard, fewer from a short
// library without losing the game (CR 701.17b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f6758e8f-5462-48a7-96d3-60d07b1563fc",
		Name:            "Scarblade Scout",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Scarblade Scout — mill two cards", effects.Do(effects.MillCards{N: 2})),
		},
	})
}
