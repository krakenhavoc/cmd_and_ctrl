package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lim-Dûl's Paladin — Creature — Human Knight {2}{B}{R}, 0/3:
//
//	"Trample
//	 At the beginning of your upkeep, you may discard a card. If you don't,
//	 sacrifice this creature and draw a card.
//	 Whenever this creature becomes blocked, it gets +6/+3 until end of turn.
//	 Whenever this creature attacks and isn't blocked, it assigns no combat
//	 damage this turn and defending player loses 4 life."
//
// #3531: the upkeep clause is a "you may" with the discard on yes; declining
// or having no card to discard is "if you don't", which sacrifices the
// Paladin and draws a card (the draw happens even if the Paladin is already
// gone). The unblocked clause has no "may": both halves happen.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c4161335-1150-4f0d-bd2b-9211efb3bea6",
		Name:            "Lim-Dûl's Paladin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Lim-Dûl's Paladin — you may discard a card; if you don't, sacrifice it and draw a card", limDulsPaladinUpkeep),
			effects.On(game.EventBecomesBlocked, effects.Self, "Lim-Dûl's Paladin — it gets +6/+3 until end of turn",
				effects.ThisGetsUntilEndOfTurn(6, 3, "Lim-Dûl's Paladin — +6/+3")),
			effects.WhenAttacksAndIsNotBlockedEffect("Lim-Dûl's Paladin — it assigns no combat damage this turn and defending player loses 4 life",
				func(g *game.Game, item *game.StackItem, defender uuid.UUID) error {
					effects.AssignsNoCombatDamageThisTurn(effects.NewContext(g, item), "Lim-Dûl's Paladin — assigns no combat damage this turn")
					if !effects.DefendingPlayerStillIn(g, defender) {
						return nil
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, defender, -4)
				}),
		},
	})
}

func limDulsPaladinUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	p := g.PlayerByIDForEffect(ctx.Controller())
	if p == nil || p.Hand.Size() == 0 {
		return limDulsPaladinSacrificeAndDraw(ctx)
	}
	return effects.MayChoice{
		Question: "Lim-Dûl's Paladin — discard a card? (if you don't, sacrifice it and draw a card)",
		Then:     limDulsPaladinAnswer,
	}.Apply(ctx)
}

// limDulsPaladinSacrificeAndDraw is "sacrifice it and draw a card".
func limDulsPaladinSacrificeAndDraw(ctx *effects.Context) error {
	if err := (effects.SacrificePermanent{Target: ctx.Source()}).Apply(ctx); err != nil {
		return err
	}
	return effects.DrawCards{N: 1}.Apply(ctx)
}

// limDulsPaladinAnswer is the upkeep choice: discard a card on a yes,
// sacrifice the Paladin and draw on a no. Keyed (#4291).
var limDulsPaladinAnswer = effects.MayYesNo("confirm/lim-duls-paladin-discard-or-sacrifice",
	func(ctx *effects.Context, _ game.ConfirmAnswer) error {
		ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: ctx.Controller(), Source: ctx.Source(), N: 1})
		return nil
	},
	func(ctx *effects.Context, _ game.ConfirmAnswer) error { return limDulsPaladinSacrificeAndDraw(ctx) })
