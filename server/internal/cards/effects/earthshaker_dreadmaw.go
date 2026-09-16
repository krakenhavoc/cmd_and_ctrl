package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Earthshaker Dreadmaw — Creature — Dinosaur {4}{G}{G}, 6/6 (EDHREC
// rank 2376):
//
//	"Trample
//	 When this creature enters, draw a card for each other Dinosaur
//	 you control."
//
// The Dinosaur deck's Harmonize on a body. The count is read as the
// trigger RESOLVES — the other Dinosaurs the controller controls
// then, post-layer subtype so a changeling counts, the Dreadmaw
// itself excluded — which is when the printed clause is evaluated.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "454cbc6a-b7f0-445e-842c-5db267917a18",
		Name:            "Earthshaker Dreadmaw",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Earthshaker Dreadmaw — draw a card for each other Dinosaur you control", func(g *game.Game, item *game.StackItem) error {
				n := b22OtherCreaturesOfSubtypeControlled(g, item.Controller, item.SourceCardID, "Dinosaur")
				return DrawCards{Player: item.Controller, N: n}.Apply(NewContext(g, item))
			}),
		},
	})
}
