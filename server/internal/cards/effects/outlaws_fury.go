package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Outlaws' Fury — Instant {2}{R} (Outlaws of Thunder Junction, #4468):
//
//	"Creatures you control get +2/+0 until end of turn. If you control
//	 an outlaw, exile the top card of your library. Until the end of
//	 your next turn, you may play that card. (Assassins, Mercenaries,
//	 Pirates, Rogues, and Warlocks are outlaws.)"
//
// The pump covers the creatures you control as the spell resolves
// (CR 611.2c), so a creature that enters later is not pumped. The outlaw
// is checked after the pump, as the spell resolves: any permanent you
// control with one of the five outlaw types counts (Mine Raider's
// reading). The exile is Crimson Operative's ExileTopNUntilYourNextTurn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5b5ef88c-f863-4f55-91fc-a63f7f3c1b3f",
		Name:         "Outlaws' Fury",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (BoostUntilEOT{
				Match: And(Creature(), YouControl()), Power: 2,
				Label: "Outlaws' Fury — +2/+0",
			}).Apply(ctx); err != nil {
				return err
			}
			if !b40ControlsA(ctx.Game, ctx.Controller(), outlaw()) {
				return nil
			}
			return ExileTopNUntilYourNextTurn(ctx, 1)
		},
	})
}
