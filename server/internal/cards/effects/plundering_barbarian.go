package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Plundering Barbarian — Creature — Dwarf Barbarian {2}{R}, 2/2 (AFR):
//
//	"When this creature enters, choose one —
//	 • Smash the Chest — Destroy target artifact.
//	 • Pry It Open — Create a Treasure token."
func init() {
	barbarian := WhenThisEnters("Plundering Barbarian — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	barbarian.Modes = ChooseOne(
		ModeDoing("Smash the Chest — Destroy target artifact.",
			TargetPermanent("target artifact", Artifact()),
			DestroyTheModesTarget),
		ModeWithPurpose(ModeDoing("Pry It Open — Create a Treasure token.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}.Apply(ctx)
			}), game.Purpose{Tokens: 1}),
	)
	Register(Spec{
		OracleID:     "da2d27f7-cced-4118-bddc-41712e37e916",
		Name:         "Plundering Barbarian",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{barbarian},
	})
}
