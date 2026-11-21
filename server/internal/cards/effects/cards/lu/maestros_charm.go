package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Maestros Charm — Instant {U}{B}{R} (Streets of New Capenna):
//
//	"Choose one —
//	 • Look at the top five cards of your library. Put one of those cards
//	   into your hand and the rest into your graveyard.
//	 • Each opponent loses 3 life and you gain 3 life.
//	 • Maestros Charm deals 5 damage to target creature or planeswalker."
//
// The look is Commune with Evil's private look-and-take (five cards, one
// to the hand, the rest put into the graveyard, not milled). The drain
// makes each opponent lose 3 life, then you gain 3; the gain happens
// whatever the opponents' life changes were, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b207754-9172-426a-8efc-a9c3912d94bb",
		Name:         "Maestros Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Look at the top five cards of your library. Put one of those cards into your hand and the rest into your graveyard.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					player := ctx.Controller()
					return effects.TakeFromLibraryToHand{
						Player: player,
						Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 5),
						Max:    1,
						Label:  "Maestros Charm — put one of those cards into your hand and the rest into your graveyard",
						Key:    effects.TakeRestGraveyard,
					}.Apply(ctx)
				}), game.Purpose{Draws: 1}),
			effects.ModeDoing("Each opponent loses 3 life and you gain 3 life.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					if err := effects.EachOpponentLosesLifeNow(ctx.Game, item, 3); err != nil {
						return err
					}
					return effects.GainLife{Player: item.Controller, Amount: 3}.Apply(ctx)
				}),
			effects.ModeDoing("Maestros Charm deals 5 damage to target creature or planeswalker.",
				effects.Harmful(effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker()))),
				effects.DealFixedDamageToModesTarget(5)),
		),
	})
}
