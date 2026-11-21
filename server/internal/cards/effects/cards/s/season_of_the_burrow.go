package cardss

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Season of the Burrow — Sorcery {3}{W}{W} (Bloomburrow, #4467):
//
//	"Choose up to five {P} worth of modes. You may choose the same mode
//	 more than once.
//	 {P} — Create a 1/1 white Rabbit creature token.
//	 {P}{P} — Exile target nonland permanent. Its controller draws a card.
//	 {P}{P}{P} — Return target permanent card with mana value 3 or less
//	 from your graveyard to the battlefield with an indestructible counter
//	 on it."
//
// A pawprint spell (CR 700.2i, #4786); each occurrence of a repeated
// targeted mode has its own target (CR 700.2d). "Its controller draws a
// card" is not conditional on the exile, so the draw is the exile's
// tail (it runs once the move is settled, whatever a replacement did)
// and the drawer is the permanent's controller as the mode resolved.
// The returned card enters under its owner's control, which is you: it
// came from your graveyard.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8e013e51-1684-4f58-8fe1-3588b6fbb5b5",
		Name:         "Season of the Burrow",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseUpToPawprints(5, true,
			effects.Pawprint(1, effects.ModeDoing("Create a 1/1 white Rabbit creature token.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.CreateToken{Template: effects.TokenCard("1/1 white Rabbit"), N: 1}.Apply(ctx)
				})),
			effects.Pawprint(2, effects.ModeDoing("Exile target nonland permanent. Its controller draws a card.",
				effects.TargetPermanent("target nonland permanent", effects.Nonland()),
				func(item *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					c, ok := ctx.Game.LookupCardForEffect(t.ID)
					if !ok {
						return nil
					}
					return effects.ExileTarget{
						Target:   t.ID,
						ThenTail: game.KeyedTail{Ref: effects.ThenDrawCards, Player: c.Controller, Source: item.SourceCardID, N: 1},
					}.Apply(ctx)
				})),
			effects.Pawprint(3, effects.ModeDoing("Return target permanent card with mana value 3 or less from your graveyard to the battlefield with an indestructible counter on it.",
				effects.TargetCardInGraveyard("target permanent card with mana value 3 or less from your graveyard", effects.YouOwn(), effects.Permanent(), effects.ManaValueLE(3)),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					_, err := ctx.Game.ReturnFromGraveyardWithCountersForEffect(t.ID, uuid.Nil, false,
						map[string]int{game.CounterIndestructible: 1})
					if errors.Is(err, game.ErrCardNotFound) {
						return nil
					}
					return err
				})),
		),
	})
}
