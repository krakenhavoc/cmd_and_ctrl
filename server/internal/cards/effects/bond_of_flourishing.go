package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "e9c9e730-3b73-48e3-845b-1b7f2d16050c",
		Name:         "Bond of Flourishing",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 3},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				player := ctx.Controller()
				return TakeFromLibraryToHand{
					Player:   player,
					Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 3),
					Match:    Permanent(),
					Max:      1,
					Optional: true,
					Label:    "Bond of Flourishing — reveal a permanent card and put it into your hand",
					Key:      takeRevealRestBottomAny,
				}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error { return GainLife{Player: ctx.Controller(), Amount: 3}.Apply(ctx) }),
		),
	})
}
