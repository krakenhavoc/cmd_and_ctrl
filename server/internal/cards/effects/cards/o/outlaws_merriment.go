package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Outlaws' Merriment — Enchantment {1}{R}{W}{W}:
//
//	"At the beginning of your upkeep, choose one at random. Create a red
//	 and white creature token with those characteristics.
//	 • 3/1 Human Warrior with trample and haste.
//	 • 2/1 Human Cleric with lifelink and haste.
//	 • 1/2 Human Rogue with haste and "When this token enters, it deals 1
//	   damage to any target.""
//
// ChooseOneAtRandom (CR 700.2): the engine draws the bullet as the
// trigger goes on the stack, from the seeded random stream. Each bullet
// creates one red and white token with the printed characteristics; the
// Rogue's enters trigger is on the token, so its target is chosen as it
// enters.
//
// No simplification.
func init() {
	t := effects.AtYourUpkeep("Outlaws' Merriment — choose one at random: a Warrior, a Cleric or a Rogue",
		func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = effects.ChooseOneAtRandom(
		effects.ModeDoing("3/1 Human Warrior with trample and haste.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("3/1 red and white Human Warrior with trample and haste"), N: 1}.Apply(ctx)
			}),
		effects.ModeDoing("2/1 Human Cleric with lifelink and haste.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("2/1 red and white Human Cleric with lifelink and haste"), N: 1}.Apply(ctx)
			}),
		effects.ModeDoing("1/2 Human Rogue with haste and \"When this token enters, it deals 1 damage to any target.\"", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.TokenFromCatalog(effects.PrintedBnsHumanRogueToken), N: 1}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "78826624-8784-4003-8813-230f299a588f",
		Name:         "Outlaws' Merriment",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
