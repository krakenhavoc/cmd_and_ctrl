package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Khorvath's Fury — Sorcery {4}{R} (Battlebond, #5529):
//
//	For each player, choose friend or foe. Each friend discards all cards from their hand, then draws that many cards plus one. Khorvath's Fury deals damage to each foe equal to the number of cards in their hand.
//
// effects.NewS82OldB205FriendOrFoe asks the caster about every player. Friends' hands are settled first, then each foe is dealt damage equal to the cards in their hand as it stands.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c3575cf9-847a-472c-94b3-ef70712c297d",
		Name:         "Khorvath's Fury",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return khorvathsFuryAsk(ctx)
		},
	})
}

var khorvathsFuryAsk func(ctx *effects.Context) error

func init() {
	khorvathsFuryAsk = effects.NewS82OldB205FriendOrFoe("option-pick/khorvaths-fury-friend-or-foe", "Khorvath's Fury",
		func(ctx *effects.Context, friends, foes []uuid.UUID) error {
			for _, friend := range friends {
				n, err := effects.DiscardWholeHand(ctx.Game, friend)
				if err != nil {
					return err
				}
				if err := (effects.DrawCards{Player: friend, N: n + 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return ctx.Game.DamageInstanceForEffect(func() error {
				for _, foe := range foes {
					p := ctx.Game.PlayerByIDForEffect(foe)
					if p == nil || p.Hand.Size() == 0 {
						continue
					}
					if err := (effects.DealDamage{Source: ctx.Source(), Target: foe, Amount: p.Hand.Size()}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})
		})
}
