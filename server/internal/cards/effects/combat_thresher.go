package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Combat Thresher — Artifact Creature — Construct {7}, 3/3 (The Brothers'
// War, #5656): "Prototype {2}{W} — 1/1", "Double strike", "When this
// creature enters, draw a card."
//
// Prototype is the card's second face (CR 718, game/prototype.go), built
// by the deck importer from the inset; nothing here spells it out.
// Double strike is a printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0c460bdd-460b-4312-9ed2-32f3809a7d36",
		Name:            "Combat Thresher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Combat Thresher — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
