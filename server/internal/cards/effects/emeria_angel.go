package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Emeria Angel — Creature — Angel {2}{W}{W}, 3/3 (EDHREC rank 1995):
//
//	"Flying
//	 Landfall — Whenever a land you control enters, you may create a
//	 1/1 white Bird creature token with flying."
//
// The landfall Bird factory. The trigger is the Tireless Provisioner
// shape — a land entering under the source's controller — with the
// "may" as a real prompt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ea6616e4-db8d-4905-80f8-cb0162906850",
		Name:            "Emeria Angel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsLand()
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Emeria Angel — create a 1/1 Bird with flying?"},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Emeria Angel — create a Bird",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{
							Controller: item.Controller,
							Template:   b18WhiteBirdToken(),
							N:          1,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
