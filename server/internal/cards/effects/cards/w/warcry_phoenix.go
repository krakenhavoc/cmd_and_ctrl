package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Warcry Phoenix — Creature — Phoenix {3}{R}, 3/3 (Jumpstart 2022, #5948):
//
//	"Flying, haste
//	 Whenever you attack with three or more creatures, you may pay {2}{R}.
//	 If you do, return this card from your graveyard to the battlefield
//	 tapped and attacking."
//
// Meddling Youths' "attack with three or more creatures" batch count, watched
// from the graveyard (the Phoenix is not on the battlefield), then Alesha's
// pay-and-return-attacking. The card enters tapped and attacking without
// having been declared as an attacker (CR 508.4), so it triggers nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5297f646-8235-47da-a290-a6883b374977",
		Name:            "Warcry Phoenix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.WithAtBatchEnd(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source != nil && effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Warcry Phoenix — you may pay {2}{R}; if you do, return this card from your graveyard to the battlefield tapped and attacking",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
						return nil
					}
					return effects.MayPay{
						Chooser:  ctx.Controller(),
						Cost:     "{2}{R}",
						Question: "Warcry Phoenix — pay {2}{R} to return it to the battlefield tapped and attacking?",
						Then:     warcryPhoenixPaid,
						Carry:    []uuid.UUID{item.SourceCardID},
					}.Apply(ctx)
				}),
				effects.ThatManyEventsAtLeast(3))),
		},
	})
}

var warcryPhoenixPaid = effects.PayUnlessBranches("pay-unless/warcry-phoenix-return",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) != 1 {
			return nil
		}
		return effects.PutAttackingYourChoice(ctx, r.Carry[0], game.ZoneGraveyard, "Warcry Phoenix")
	}, nil)
