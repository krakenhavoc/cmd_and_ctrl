package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "3b66d2c2-be7a-4296-9888-f0cfb2975e89",
		Name:            "Seismic Monstrosaur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []ActivatedAbility{
			{
				Label:   "{2}{R}, Sacrifice a land: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    Plus(ManaCost("{2}{R}"), SacrificeN(1, "a land", Land())),
				Effect:  Do(DrawCards{N: 1}),
			},
			Typecycling("Mountaincycling", "{2}", "a Mountain card", IsLandWithSubtype("Mountain")),
		},
	})
}
