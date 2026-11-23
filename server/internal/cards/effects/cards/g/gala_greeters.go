package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// b15GalaGreetersLabel is the trigger's stack label, and with it the
// key its "hasn't been chosen this turn" memory is kept under.
const B15GalaGreetersLabel = "Gala Greeters — alliance"

// Gala Greeters — Creature — Elf Druid {1}{G}, 1/1 (EDHREC rank
// 1706):
//
//	"Alliance — Whenever another creature you control enters, choose
//	 one that hasn't been chosen this turn —
//	 • Put a +1/+1 counter on this creature.
//	 • Create a tapped Treasure token.
//	 • You gain 2 life."
//
// A two-drop that pays a counter, a Treasure and two life for the
// first three creatures each turn. Corpse Knight's condition
// (b13AnotherCreatureYouControlEntered); the modes are each one
// primitive.
//
// #764 made it a real mode choice: TriggeredAbility.Modes is the same
// game.ModeSpec a modal spell declares, chosen as the ability is put on
// the stack (CR 603.3c) through a mode_pick prompt to the controller.
// The trigger targets nothing, so the prompt is the only thing between
// the harvest and the stack.
//
// ADR 0097 (#1749) made "that hasn't been chosen this turn" real too:
// ChooseOneNotChosenThisTurn declares it, and the engine remembers the
// modes this Greeters' ability has chosen this turn, recorded as each
// is chosen. Four creatures entering at once queue four prompts; each
// answer takes its bullet off the others, and the fourth is withdrawn
// with nothing left to choose — the 2022-04-29 ruling's "that choice
// is made only for the first three". A Greeters that leaves and
// returns is a new object and may choose all three again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cce081eb-8820-415a-a7b2-3c5b9d4a2601",
		Name:         "Gala Greeters",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			B15GalaGreetersTrigger(),
		},
	})
}

// b15GalaGreetersTrigger is the alliance trigger with its three
// bullets. Each bullet declares its own body (ModeDoing), so the
// engine dispatches the chosen one at resolution in announce order
// (CR 608.2c) and the card file writes no switch.
func B15GalaGreetersTrigger() game.TriggeredAbility {
	t := effects.WheneverAnotherCreatureEntersUnderYourControl(B15GalaGreetersLabel,
		func(g *game.Game, item *game.StackItem) error { return nil })
	t.Modes = effects.ChooseOneNotChosenThisTurn(
		effects.ModeDoing("Put a +1/+1 counter on this creature.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				if !effects.B15OnBattlefield(ctx.Game, item.SourceCardID) {
					return nil
				}
				return effects.AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 1}.Apply(ctx)
			}),
		effects.ModeDoing("Create a tapped Treasure token.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.B13CreateTappedTreasures(ctx, item.Controller, 1)
			}),
		effects.ModeDoing("You gain 2 life.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.GainLife{Player: item.Controller, Amount: 2}.Apply(ctx)
			}),
	)
	return t
}
