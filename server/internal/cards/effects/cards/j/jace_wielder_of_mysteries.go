package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jace, Wielder of Mysteries — Legendary Planeswalker — Jace {1}{U}{U}{U},
// starting loyalty 4:
//
//	"If you would draw a card while your library has no cards in it, you win
//	 the game instead.
//	 +1: Target player mills two cards. Draw a card.
//	 −8: Draw seven cards. Then if your library has no cards in it, you win
//	 the game."
//
// The static is Laboratory Maniac's draw replacement
// (WinInsteadOfDrawingFromAnEmptyLibrary). The +1 mills the target, then
// the controller draws once the mill has finished (CR 608.2c). The −8 draws
// seven (each draw an empty library replaces wins), then checks the library
// as it is now.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5277a68e-1ceb-4b40-8e10-3d0564f5d7be",
		Name:            "Jace, Wielder of Mysteries",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Replacements:    []game.ReplacementEffect{effects.WinInsteadOfDrawingFromAnEmptyLibrary("Jace, Wielder of Mysteries")},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Target player mills two cards. Draw a card.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetPlayer("target player"),
				Effect: effects.Do(
					effects.Instruction(func(ctx *effects.Context) error { return effects.JacePerfectedMindMill(ctx, 2) }),
					effects.DrawCards{N: 1},
				),
			},
			{
				Label:  "−8: Draw seven cards. Then if your library has no cards in it, you win the game.",
				Cost:   effects.LoyaltyCost(-8),
				Effect: effects.Do(effects.DrawCards{N: 7}, effects.Instruction(jaceWielderOfMysteriesWin)),
			},
		},
	})
}

// jaceWielderOfMysteriesWin is the −8's second sentence.
func jaceWielderOfMysteriesWin(ctx *effects.Context) error {
	p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
	if p == nil || p.Library == nil || p.Library.Size() > 0 {
		return nil
	}
	_, err := ctx.Game.WinTheGameForEffect(ctx.Controller(), ctx.Item.SourceCardID)
	return err
}
