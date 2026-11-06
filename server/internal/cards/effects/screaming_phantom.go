package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Screaming Phantom — Creature — Spirit {2}{B}, 2/2 (The Lost Caverns of
// Ixalan, #4470):
//
//	"Flying
//	 Whenever this creature attacks, mill a card."
//
// Flying is the printed keyword (CR 702.9a). The attack trigger mills
// its controller's top card (CR 701.17a); it is not optional, unlike
// Hedge Shredder's, and a library with no card mills nothing and costs
// nobody the game.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d32f570a-3738-4a9d-845f-6c7892b69991",
		Name:            "Screaming Phantom",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Screaming Phantom — mill a card", Do(MillCards{N: 1})),
		},
	})
}
