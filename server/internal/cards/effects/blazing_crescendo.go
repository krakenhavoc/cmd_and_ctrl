package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blazing Crescendo — Instant {1}{R}:
//
//	"Target creature gets +3/+1 until end of turn.
//	 Exile the top card of your library. Until the end of your next
//	 turn, you may play that card."
//
// The boost is Titan's Strength's; the exile is the "until the end of
// your next turn" impulse (ExileTopNUntilYourNextTurn), which grants a
// play permission rather than a cast-only one, so an exiled land can
// be played.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6344c96a-efa5-4125-8219-d333228391cf",
		Name:         "Blazing Crescendo",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard {
				if err := (BoostUntilEOT{
					Target: item.Targets[0].ID,
					Power:  3, Toughness: 1,
					Label: "Blazing Crescendo — +3/+1 until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return ExileTopNUntilYourNextTurn(ctx, 1)
		},
	})
}
