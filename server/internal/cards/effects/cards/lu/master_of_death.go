package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Master of Death — Creature — Zombie Wizard {1}{U}{B}, 3/1:
//
//	"When this creature enters, surveil 2.
//	 At the beginning of your upkeep, if this card is in your graveyard,
//	 you may pay 1 life. If you do, return it to your hand."
//
// The upkeep trigger watches from the graveyard only (InGraveyard,
// CR 113.6, 108.4: "you" is the card's owner there). Its intervening if
// (CR 603.4) is that zone, so the trigger exists only while the card is
// there, and the return re-checks it on resolution. The payment is
// Corpses of the Lost's (MayPay, LifePayment): a player who can't pay
// life is not asked, and the card returns only once the life is paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a16bdda-8499-47e4-b870-b9576f4c1a83",
		Name:         "Master of Death",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Master of Death — surveil 2", effects.Do(effects.Surveil{N: 2})),
			effects.InGraveyard(effects.AtYourUpkeep("Master of Death — you may pay 1 life to return it to your hand",
				func(g *game.Game, item *game.StackItem) error {
					if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
						return nil
					}
					return effects.MayPay{
						Chooser:  item.Controller,
						Pay:      effects.LifePayment(1),
						Question: "Master of Death — pay 1 life to return it to your hand?",
						Then:     masterOfDeathPaid,
					}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}

// masterOfDeathPaid is "return it to your hand", once the life is paid.
// An on-disk key: never renamed, never reused.
var masterOfDeathPaid = effects.PayUnlessBranches("pay-unless/master-of-death-return",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.ReturnThisCardFromYourGraveyardToHand(ctx.Game, ctx.Item)
	}, nil)
