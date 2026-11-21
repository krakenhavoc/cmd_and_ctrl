package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bramble Sovereign — Creature — Dryad {2}{G}{G}, 4/4 (Battlebond, #5529):
//
//	"Whenever another nontoken creature enters, you may pay {1}{G}. If you
//	 do, that creature's controller creates a token that's a copy of that
//	 creature."
//
// Any player's nontoken creature but the Sovereign itself
// (bbdAnotherNontokenCreatureEntered). The {1}{G} is asked as the trigger
// resolves (MayPay); paid, the token is created by that creature's
// controller, read as the trigger resolves, or as it last existed on the
// battlefield if it has gone (CR 608.2h), and it copies the creature's
// copiable values the same way (CreateTokenCopy, which reads last-known
// information for a permanent that left).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e26a688-67d7-48a5-8862-653851bf3784",
		Name:         "Bramble Sovereign",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.BbdAnotherNontokenCreatureEntered,
				"Bramble Sovereign — you may pay {1}{G}; if you do, that creature's controller creates a token copy of it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}{G}",
						Question: "Bramble Sovereign — pay {1}{G} so that creature's controller creates a token that's a copy of it?",
						Then:     brambleSovereignPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// brambleSovereignPaid is the paid branch. An on-disk key: never renamed,
// never reused.
var brambleSovereignPaid = effects.PayUnlessBranches("pay-unless/bramble-sovereign-copy",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		info, ok := ctx.TriggeringPermanent()
		if !ok || info.Controller == uuid.Nil {
			return nil
		}
		return effects.CreateTokenCopy{Controller: info.Controller, Copy: ctx.Trigger().Event.CardID, N: 1}.Apply(ctx)
	}, nil)
