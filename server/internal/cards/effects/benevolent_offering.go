package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Benevolent Offering — Instant {3}{W}:
//
//	"Choose an opponent. You and that player each create three 1/1 white
//	 Spirit creature tokens with flying.
//	 Choose an opponent. You gain 2 life for each creature you control and
//	 that player gains 2 life for each creature they control."
//
// Intellectual Offering's shape: two separate choices, so the second
// opponent may differ from the first, each a real prompt that stays a
// restore point (keyed). The creatures are counted as the second clause
// is carried out (CR 608.2c), after the tokens exist, so the first
// clause's Spirits count for whoever made them. With no opponent left to
// choose, the player half of each clause still happens for you alone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3db8466c-864f-48ba-b617-f2c863dfdcbd",
		Name:         "Benevolent Offering",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ChoosePlayer{
				Among:    Opponents,
				Question: "Benevolent Offering — choose an opponent; you and that player each create three 1/1 white Spirit creature tokens with flying",
				ThenKey:  benevolentOfferingTokens,
			}.Apply(ctx)
		},
	})
}

var benevolentOfferingTokens = ChoosePlayerThen("option-pick/benevolent-offering-tokens", func(ctx *Context, chosen uuid.UUID, _ []uuid.UUID) error {
	if err := (CreateToken{Template: b28WhiteSpiritFlyingToken(), N: 3}).Apply(ctx); err != nil {
		return err
	}
	if chosen != uuid.Nil {
		if err := (CreateToken{Controller: chosen, Template: b28WhiteSpiritFlyingToken(), N: 3}).Apply(ctx); err != nil {
			return err
		}
	}
	return ChoosePlayer{
		Among:    Opponents,
		Question: "Benevolent Offering — choose an opponent; you gain 2 life for each creature you control and that player gains 2 life for each creature they control",
		ThenKey:  benevolentOfferingLife,
	}.Apply(ctx)
})

var benevolentOfferingLife = ChoosePlayerThen("option-pick/benevolent-offering-life", func(ctx *Context, chosen uuid.UUID, _ []uuid.UUID) error {
	creatures := func(p uuid.UUID) int {
		n := 0
		for _, c := range ctx.Game.BattlefieldCardsForEffect() {
			if c.Controller == p && c.IsCreature() {
				n++
			}
		}
		return n
	}
	if err := (GainLife{Amount: 2 * creatures(ctx.Controller())}).Apply(ctx); err != nil {
		return err
	}
	if chosen == uuid.Nil {
		return nil
	}
	return GainLife{Player: chosen, Amount: 2 * creatures(chosen)}.Apply(ctx)
})
