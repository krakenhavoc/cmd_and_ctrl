package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// litaLabel is the trigger's stack label, and with it the key its
// "hasn't been chosen this turn" memory is kept under.
const litaLabel = "Lita, Little Orphan Amphibian — alliance"

// Lita, Little Orphan Amphibian — Legendary Creature — Mutant Ninja
// Turtle {1}{W}, 2/1:
//
//	"Alliance — Whenever another creature you control enters, choose
//	 one that hasn't been chosen this turn.
//	 • Put a +1/+1 counter on Lita.
//	 • Create a Food token.
//	 • Scry 1."
//
// Gala Greeters' shape in white: the alliance condition and
// ChooseOneNotChosenThisTurn (ADR 0097). The first three creatures to
// enter each turn pay three different bullets; a fourth triggers and
// is removed with no effect.
//
// The counter goes on Lita by object — AddCounter refuses a Lita that
// left and came back before the trigger resolved (CR 400.7). The Food
// is the catalog's Food token, with its own sacrifice ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "212fdb7c-1c7e-4eed-b1ed-bcc14a425da8",
		Name:         "Lita, Little Orphan Amphibian",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			litaTrigger(),
		},
	})
}

func litaTrigger() game.TriggeredAbility {
	t := effects.WheneverAnotherCreatureEntersUnderYourControl(litaLabel,
		func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = effects.ChooseOneNotChosenThisTurn(
		effects.ModeDoing("Put a +1/+1 counter on Lita.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
		effects.ModeDoing("Create a Food token.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.FoodToken(), N: 1}.Apply(ctx)
			}),
		effects.ModeDoing("Scry 1.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.Scry{Player: item.Controller, N: 1}.Apply(ctx)
			}),
	)
	return t
}
