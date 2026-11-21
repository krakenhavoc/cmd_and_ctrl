package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plundering Barbarian — Creature — Dwarf Barbarian {2}{R}, 2/2 (AFR):
//
//	"When this creature enters, choose one —
//	 • Smash the Chest — Destroy target artifact.
//	 • Pry It Open — Create a Treasure token."
func init() {
	barbarian := effects.WhenThisEnters("Plundering Barbarian — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	barbarian.Modes = effects.ChooseOne(
		effects.ModeDoing("Smash the Chest — Destroy target artifact.",
			effects.TargetPermanent("target artifact", effects.Artifact()),
			effects.DestroyTheModesTarget),
		effects.ModeWithPurpose(effects.ModeDoing("Pry It Open — Create a Treasure token.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.TreasureToken(), N: 1}.Apply(ctx)
			}), game.Purpose{Tokens: 1}),
	)
	effects.Register(effects.Spec{
		OracleID:     "da2d27f7-cced-4118-bddc-41712e37e916",
		Name:         "Plundering Barbarian",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{barbarian},
	})
}
