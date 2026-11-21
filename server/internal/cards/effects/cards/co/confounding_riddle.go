package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Confounding Riddle — Instant {2}{U} (The Lost Caverns of Ixalan, #4470):
//
//	"Choose one —
//	 • Look at the top four cards of your library. Put one of them into
//	   your hand and the rest into your graveyard.
//	 • Counter target spell unless its controller pays {4}."
//
// The first mode is Commune with Evil's look: a private look at the top
// four, a mandatory pick of one for the hand, and the rest put into the
// graveyard (takeRestGraveyard), not milled or discarded. A shorter
// library offers what it has. The second is Confusticate and Bebother's
// CounterUnlessPaid: the spell's controller may pay {4} as the mode
// resolves (CR 701.6a); if they don't, the spell is countered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3df49c4d-a285-4826-824e-b742b0deed3a",
		Name:         "Confounding Riddle",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Look at the top four cards of your library. Put one of them into your hand and the rest into your graveyard.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					player := ctx.Controller()
					return effects.TakeFromLibraryToHand{
						Player: player,
						Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 4),
						Max:    1,
						Label:  "Confounding Riddle — put one of them into your hand and the rest into your graveyard",
						Key:    effects.TakeRestGraveyard,
					}.Apply(ctx)
				}),
				game.Purpose{Draws: 1}),
			effects.ModeDoing("Counter target spell unless its controller pays {4}.",
				effects.TargetSpell("target spell"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.CounterUnlessPaid{
						StackID:  t.ID,
						Cost:     "{4}",
						Question: "Confounding Riddle — pay {4} or your spell is countered",
					}.Apply(ctx)
				}),
		),
	})
}
