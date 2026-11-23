package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "6156bd59-d623-4612-8e86-b5228d5de750",
		Name:            "Howling Banshee",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Howling Banshee — each player loses 3 life", effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					for _, id := range effects.TablePlayers(ctx) {
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
