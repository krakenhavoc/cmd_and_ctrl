package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thought Monitor — Artifact Creature — Construct {6}{U}, 2/2:
//
//	"Affinity for artifacts (This spell costs {1} less to cast for
//	 each artifact you control.)
//	 Flying
//	 When this creature enters, draw two cards."
//
// #746: affinity is a spell's own cost modifier (CR 702.41a), read
// from Spec.SelfCostModifiers while the card is being cast.
func init() {
	Register(Spec{
		OracleID:        "9deded8b-cec4-4ede-a50b-131404d456d4",
		Name:            "Thought Monitor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for artifacts", Artifact()),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Thought Monitor — draw two cards", Do(DrawCards{N: 2})),
		},
	})
}
