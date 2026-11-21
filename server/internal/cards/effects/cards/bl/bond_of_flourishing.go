package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bond of Flourishing — Sorcery {1}{G} (Foundations Jumpstart):
//
//	"Look at the top three cards of your library. You may reveal a
//	 permanent card from among them and put it into your hand. Put the
//	 rest on the bottom of your library in any order. You gain 3 life."
//
// Commune with Nature's dig over permanent cards, then the life: it is
// gained whether or not a card was taken.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9c9e730-3b73-48e3-845b-1b7f2d16050c",
		Name:         "Bond of Flourishing",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 3},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				player := ctx.Controller()
				return effects.TakeFromLibraryToHand{
					Player:   player,
					Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 3),
					Match:    effects.Permanent(),
					Max:      1,
					Optional: true,
					Label:    "Bond of Flourishing — reveal a permanent card and put it into your hand",
					Key:      effects.TakeRevealRestBottomAny,
				}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.GainLife{Player: ctx.Controller(), Amount: 3}.Apply(ctx)
			}),
		),
	})
}
