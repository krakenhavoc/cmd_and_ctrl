package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Acolyte of Affliction — Creature — Human Cleric {2}{B}{G}, 2/3:
//
//	"When this creature enters, mill two cards, then you may return a
//	 permanent card from your graveyard to your hand."
//
// The choice is made after the mill, so a milled permanent card is a legal
// pick.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d57c246f-18e5-4523-898c-1509db738363",
		Name:         "Acolyte of Affliction",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Acolyte of Affliction — mill two cards, then you may return a permanent card from your graveyard to your hand",
				effects.Do(
					effects.Instruction(func(ctx *effects.Context) error { return effects.MillCards{N: 2}.Apply(ctx) }),
					effects.Instruction(func(ctx *effects.Context) error {
						return effects.ReturnChosenFromGraveyard{
							Player:   ctx.Controller(),
							Question: "Acolyte of Affliction — you may return a permanent card from your graveyard to your hand",
							Match:    game.Card.IsPermanent,
							Max:      1,
							ToHand:   true,
						}.Apply(ctx)
					}))),
		},
	})
}
