package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hugs, Grisly Guardian — Legendary Creature — Badger Warrior {X}{R}{R}{G}{G}, 5/5 (Bloomburrow):
//
//	"Trample
//	 When Hugs enters, exile the top X cards of your library. Until the end of
//	 your next turn, you may play those cards.
//	 You may play an additional land on each of your turns."
//
// X is the X the spell was cast with, read by the enters trigger (ctx.X(), as
// Alquist Proft does); the exile and the play window are Atsushi's (ExileTopNUntilYourNextTurn:
// the cards are played with their own costs and timing, a land in a main phase
// with the stack empty, even if Hugs has left). The extra land drop is the
// Exploration clause (Spec.AdditionalLandPlays), cumulative with others.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:            "3c5c227a-335b-4995-9821-d4269a0a233a",
		Name:                "Hugs, Grisly Guardian",
		XMatters:            true,
		Completeness:        effects.CompletenessFull,
		PrintedKeywords:     []string{"trample"},
		AdditionalLandPlays: 1,
		Purpose:             game.Purpose{ExtraLandDrops: 1},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.Self,
			Key:       "Hugs, Grisly Guardian — exile the top X cards of your library; until the end of your next turn, you may play those cards",
			// X is the X the spell was cast with (CR 107.3m).
			XFromCast: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.ExileTopNUntilYourNextTurn(ctx, ctx.X())
			},
		}},
	})
}
