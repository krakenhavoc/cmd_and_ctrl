package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "a4ffe297-5e82-43f9-91a4-7aa3d8dd3b4a",
		Name:            "Crow of Dark Tidings",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrDies("Crow of Dark Tidings — mill two cards", Do(MillCards{N: 2})),
		},
	})
}
