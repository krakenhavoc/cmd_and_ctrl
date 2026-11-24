package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Song of Creation — Enchantment {1}{G}{U}{R} (Secret Lair Drop):
//
//	"You may play an additional land on each of your turns.
//	 Whenever you cast a spell, draw two cards.
//	 At the beginning of your end step, discard your hand."
//
// The extra land play is the printed AdditionalLandPlays (Zell Dincht's), one
// for each of its controller's turns. Every spell cast draws two, resolved
// before the spell. The end-step discard is the whole hand, a hand of no cards
// included, at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:            "75b73eb2-dafd-43db-9e96-7e1995d9849f",
		Name:                "Song of Creation",
		Completeness:        effects.CompletenessFull,
		AdditionalLandPlays: 1,
		Purpose:             game.Purpose{ExtraLandDrops: 1, Draws: 2},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(nil, "Song of Creation — draw two cards", effects.Do(effects.DrawCards{N: 2})),
			effects.AtYourEndStep("Song of Creation — discard your hand",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					_, err := effects.DiscardWholeHand(ctx.Game, ctx.Controller())
					return err
				}))),
		},
	})
}
