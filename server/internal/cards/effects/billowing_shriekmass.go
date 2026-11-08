package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Billowing Shriekmass — Creature — Spirit {3}{B}, 2/3 (Foundations,
// #4465):
//
//	"Flying
//	 When this creature enters, mill three cards.
//	 Threshold — This creature gets +2/+1 as long as there are seven or
//	 more cards in your graveyard."
//
// Dark Matter Manipulator's enters mill, and Most Decrepit Old Bird's
// threshold static (fraThresholdSelfPT), read live on every recompute.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ce45d6bd-74e3-401b-983d-028d3fb942da",
		Name:            "Billowing Shriekmass",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Billowing Shriekmass — mill three cards", Do(MillCards{N: 3})),
		},
		Static: []game.StaticAbility{fraThresholdSelfPT(2, 1)},
	})
}
