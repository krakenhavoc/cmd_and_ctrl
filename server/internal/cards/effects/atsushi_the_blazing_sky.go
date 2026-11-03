package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Atsushi, the Blazing Sky — Legendary Creature — Dragon Spirit {2}{R}{R},
// 4/4:
//
//	"Flying, trample
//	 When Atsushi dies, choose one —
//	 • Exile the top two cards of your library. Until the end of your
//	   next turn, you may play those cards.
//	 • Create three Treasure tokens."
//
// Ao, the Dawn Sky's modal dies trigger: the mode is chosen as the trigger
// goes on the stack (CR 603.3c) and each bullet declares its own body. The
// first is Reckless Impulse's exile (ExileTopNUntilYourNextTurn, ADR 0063's
// "your next turn" window); the second is three Treasures.
//
// No simplification.
func init() {
	dies := WhenThisDies("Atsushi, the Blazing Sky — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	dies.Modes = ChooseOne(
		ModeDoing("Exile the top two cards of your library. Until the end of your next turn, you may play those cards.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return ExileTopNUntilYourNextTurn(ctx, 2)
			}),
		ModeDoing("Create three Treasure tokens.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return CreateToken{Controller: ctx.Controller(), Template: TreasureToken(), N: 3}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:        "535f405d-94df-4f45-aedc-b331103720b7",
		Name:            "Atsushi, the Blazing Sky",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered:       []game.TriggeredAbility{dies},
	})
}
