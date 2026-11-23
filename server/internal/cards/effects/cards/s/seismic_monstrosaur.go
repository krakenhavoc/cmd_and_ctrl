package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seismic Monstrosaur — Creature — Dinosaur {4}{R}{R}, 6/5 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Trample
//	 {2}{R}, Sacrifice a land: Draw a card.
//	 Mountaincycling {2} ({2}, Discard this card: Search your library for
//	 a Mountain card, reveal it, put it into your hand, then shuffle.)"
//
// Trample is the printed keyword (CR 702.19a). The draw sacrifices a land
// its controller controls as part of the cost (Dal R'El, Augment Orphan's
// row with {2}{R}). Mountaincycling is the shared Typecycling constructor
// (CR 702.29e) with "a Mountain card": any land with the Mountain type,
// as Oliphaunt has it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3b66d2c2-be7a-4296-9888-f0cfb2975e89",
		Name:            "Seismic Monstrosaur",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}{R}, Sacrifice a land: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.SacrificeN(1, "a land", effects.Land())),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
			effects.Typecycling("Mountaincycling", "{2}", "a Mountain card", effects.IsLandWithSubtype("Mountain")),
		},
	})
}
