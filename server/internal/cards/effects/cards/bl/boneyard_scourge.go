package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Boneyard Scourge — Creature — Zombie Dragon {2}{B}{B}, 4/3 (Commander
// 2017, #5533):
//
//	"Flying
//	 Whenever a Dragon you control dies while this card is in your
//	 graveyard, you may pay {1}{B}. If you do, return this card from your
//	 graveyard to the battlefield."
//
// Furious Forebear's graveyard trigger (CR 113.6, InGraveyard) narrowed to a
// Dragon you control, read from last-known information (CR 603.10a), with
// the return going to the battlefield. The card must already have been in
// the graveyard, so the Scourge's own death, or one in the batch that put it
// there, does not trigger it. "You may pay {1}{B}. If you do" is a MayPay
// asked as the trigger resolves, and the return needs the card to still be
// in the graveyard as the same object (CR 400.7).
//
// No simplification.
func init() {
	t := effects.On(game.EventLTB, effects.C17ADragonYouControlDiedWhileInGraveyard,
		"Boneyard Scourge — you may pay {1}{B}; if you do, return this card from your graveyard to the battlefield",
		func(g *game.Game, item *game.StackItem) error {
			return effects.MayPay{
				Chooser:  item.Controller,
				Cost:     "{1}{B}",
				Question: "Boneyard Scourge — pay {1}{B} to return it from your graveyard to the battlefield?",
				Then:     boneyardScourgePaid,
			}.Apply(effects.NewContext(g, item))
		})
	effects.Register(effects.Spec{
		OracleID:        "eef40ab2-9e17-4e8d-9e58-1a82ab871b5b",
		Name:            "Boneyard Scourge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{effects.InGraveyard(t)},
	})
}

// boneyardScourgePaid returns the card to the battlefield if it is still in
// the graveyard as the object the ability came from. An on-disk key.
var boneyardScourgePaid = effects.PayUnlessBranches("pay-unless/boneyard-scourge-return",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		self := ctx.Source()
		z := ctx.Game.FindCardZoneForEffect(self)
		if self == uuid.Nil || z == nil || z.Kind != game.ZoneGraveyard {
			return nil
		}
		card, ok := ctx.Game.LookupCardForEffect(self)
		if !ok {
			return nil
		}
		if ref := ctx.Item.SourceObject; ref.ID == self && ref.Epoch != card.ObjectEpoch {
			return nil
		}
		return effects.ReturnFromGraveyard{Target: self, Dest: game.ZoneBattlefield}.Apply(ctx)
	}, nil)
