package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Memory Deluge — Instant {2}{U}{U}:
//
//	"Look at the top X cards of your library, where X is the amount of
//	 mana spent to cast this spell. Put two of them into your hand and
//	 the rest on the bottom of your library in a random order.
//	 Flashback {5}{U}{U}"
//
// X is the payment record's Total (game.ManaSpent): four from hand,
// seven from the graveyard, and more when a tax or a cost increase was
// paid on top. A copy spent nothing (CR 707.10) and looks at no cards.
// A payment the engine waived (strict mana off) also reads as nothing,
// which is the weaker-than-printed direction; the caveat says so, the
// way Pentad Prism's does.
//
// The take is Dig Through Time's mandatory two, with the rest going to
// the bottom in a random order rather than an ordered one.
func init() {
	Register(Spec{
		OracleID:     "e6fd55f2-7e26-469c-a44a-ea2eb90e19a9",
		Name:         "Memory Deluge",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"With strict mana off, the game doesn't track the mana you spent, so it looks at no cards.",
		},
		Purpose:          game.Purpose{Tutors: 2},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{5}{U}{U}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			x := ctx.ManaSpent().Total()
			if x <= 0 {
				return nil
			}
			return TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, x),
				Max:    2,
				Label:  "Memory Deluge — put two of them into your hand",
				Then:   TakeRestOnBottomInRandomOrder,
			}.Apply(ctx)
		},
	})
}
