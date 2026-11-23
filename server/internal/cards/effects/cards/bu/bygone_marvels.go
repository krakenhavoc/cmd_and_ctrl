package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bygone Marvels — Sorcery {G}{G} (Lost Caverns of Ixalan Commander):
//
//	"Descend 8 — When you cast this spell, if there are eight or more
//	 permanent cards in your graveyard, copy this spell twice. You may choose
//	 new targets for the copies.
//	 Return target permanent card from your graveyard to your hand. Exile
//	 Bygone Marvels."
//
// The cast trigger carries an intervening "if" (CR 603.4) over the permanent
// cards in your graveyard (CondDescend, checked as the spell is cast and
// again as the trigger resolves) and makes two independent copies of the
// spell (CopySpell Count 2, each offered its own new target; a spell named,
// not targeted, so FromLastKnown keeps the copies when the spell was
// countered in response). The spell returns its target to your hand and then
// exiles itself; a copy is not a card, so it only returns its target and
// ceases to exist (CR 707.10a).
//
// No simplification.
func init() {
	trigger := effects.WhenYouCastThisSpell("Bygone Marvels — descend 8: copy this spell twice; you may choose new targets for the copies",
		func(g *game.Game, item *game.StackItem) error {
			return effects.CopySpell{StackID: item.SourceCardID, Controller: item.Controller, Count: 2,
				ChooseNewTargets: true, FromLastKnown: true}.Apply(effects.NewContext(g, item))
		})
	trigger.AppliesTo = effects.AllOf(effects.Self, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
		return effects.PermanentCardsInGraveyard(g, ev.Actor) >= 8
	})
	trigger = effects.InterveningIf(effects.CondDescend{N: 8}, trigger)
	effects.Register(effects.Spec{
		OracleID:     "494dc919-e429-491b-b19b-5037499e2dd6",
		Name:         "Bygone Marvels",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target permanent card from your graveyard", effects.YouOwn(), effects.Permanent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.ReturnLegalGraveyardTargetsToHand(ctx); err != nil {
				return err
			}
			if item.IsCopy || effects.SourceIsNewObject(ctx.Game, item) {
				return nil
			}
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
		Triggered: []game.TriggeredAbility{trigger},
	})
}
