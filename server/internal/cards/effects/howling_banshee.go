package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Howling Banshee — Creature — Spirit {2}{B}{B}, 3/3 (Foundations
// Jumpstart, BNS batch 32):
//
//	"Flying
//	 When this creature enters, each player loses 3 life."
//
// Every player, you included, loses 3 life as the trigger resolves
// (life loss, not damage), the controller first.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6156bd59-d623-4612-8e86-b5228d5de750",
		Name:            "Howling Banshee",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Howling Banshee — each player loses 3 life", Do(
				Instruction(func(ctx *Context) error {
					for _, id := range tablePlayers(ctx) {
						if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), id, -3); err != nil {
							return err
						}
					}
					return nil
				}),
			)),
		},
	})
}
