package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blood Operative — Creature — Vampire Assassin {1}{B}{B}, 3/1 (Guilds of
// Ravnica, #5525):
//
//	"Lifelink
//	 When this creature enters, you may exile target card from a graveyard.
//	 Whenever you surveil, if this card is in your graveyard, you may pay 3
//	 life. If you do, return this card to your hand."
//
// The enters trigger is Deathgorge Scavenger's optional exile of a card from any
// graveyard, with no rider. The surveil trigger is Master of Death's graveyard
// trigger (InGraveyard, CR 113.6; its intervening "if" is that zone, rechecked on
// resolution, CR 603.4) on Dimir Spybug's "whenever you surveil" event, with
// Master of Death's MayPay of life: a player who can't pay 3 life is not asked,
// and the card returns only once the life is paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ef75da10-1b33-4066-bd42-1111b5a94f75",
		Name:            "Blood Operative",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Optional(effects.WhenThisEnters("Blood Operative — exile target card from a graveyard", effects.SocB37ExileTheLegalTarget),
				"Blood Operative — exile target card from a graveyard?"),
				effects.TargetCardInGraveyard("target card from a graveyard")),
			effects.InGraveyard(effects.Cmdr2B27WheneverYouSurveil("Blood Operative — you may pay 3 life to return it to your hand",
				func(g *game.Game, item *game.StackItem) error {
					if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
						return nil
					}
					return effects.MayPay{
						Chooser:  item.Controller,
						Pay:      effects.LifePayment(3),
						Question: "Blood Operative — pay 3 life to return it to your hand?",
						Then:     bloodOperativePaid,
					}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}

// bloodOperativePaid is "return this card to your hand", once the life is paid.
// An on-disk key: never renamed, never reused.
var bloodOperativePaid = effects.PayUnlessBranches("pay-unless/blood-operative-return",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.ReturnThisCardFromYourGraveyardToHand(ctx.Game, ctx.Item)
	}, nil)
