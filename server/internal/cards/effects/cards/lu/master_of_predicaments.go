package cardslu

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Master of Predicaments — Creature — Sphinx {3}{U}{U}, 4/4 (Magic 2015,
// #5556):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, choose a card in
//	 your hand. That player guesses whether the card's mana value is greater
//	 than 4. If the player guessed wrong, you may cast the card without
//	 paying its mana cost."
//
// The Seventh Doctor's guess with the number fixed at 4 and no consolation
// prize: the controller picks a card from their hand (it stays hidden), the
// player the Master hit guesses "greater" or "not greater", and the answer
// is judged against the card's mana value (X is 0 in a hand, CR 107.3b). A
// wrong guess offers the free cast as part of the resolution (CR 608.2g);
// declining it, or a card that can't be cast (a land), does nothing. An
// empty hand asks nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "503447d9-2489-46d7-b5aa-a14bbb22dcc8",
		Name:            "Master of Predicaments",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Master of Predicaments — choose a card in your hand; that player guesses whether its mana value is greater than 4; if wrong you may cast it free",
				masterOfPredicamentsDamage),
		},
	})
}

func masterOfPredicamentsDamage(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	hand := effects.AllHandCardIDs(g, item.Controller)
	if len(hand) == 0 || item.Trigger == nil || item.Trigger.Event.Target == uuid.Nil {
		return nil
	}
	// #6175: the player guesses. Noted here, where the resolving item is
	// still in hand, and only when a guess will actually be asked.
	effects.NotePileOrGuess(ctx)
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   "Master of Predicaments — choose a card in your hand",
		Cards:      hand,
		Min:        1,
		Max:        1,
		ThenKey:    masterOfPredicamentsChosen,
		Carry:      []uuid.UUID{item.Trigger.Event.Target},
		LiveZone:   game.ZoneHand,
	})
	return nil
}

// masterOfPredicamentsChosen is the controller's pick; the player they hit is
// asked to guess. An on-disk key: never renamed, never reused.
var masterOfPredicamentsChosen = game.RegisterChooseCardsThen("choose-cards/master-of-predicaments-chosen", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || len(r.Carry) == 0 {
		return nil
	}
	ctx := effects.KeyedContext(g, r.Chooser, r.Source)
	return effects.MayChoice{
		Player:   r.Carry[0],
		Question: "Master of Predicaments — is the mana value of the card they chose greater than 4?",
		YesLabel: "Greater than 4",
		NoLabel:  "Not greater than 4",
		Then:     masterOfPredicamentsGuessed,
		Carry:    []uuid.UUID{r.Picked[0]},
	}.Apply(ctx)
})

// masterOfPredicamentsGuessed judges the guess (Accepted is "greater"). An
// on-disk key: never renamed, never reused.
var masterOfPredicamentsGuessed = effects.MayAnswer("confirm/master-of-predicaments-guess", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	g, controller, source := ctx.Game, ctx.Controller(), ctx.Source()
	if len(r.Carry) == 0 {
		return nil
	}
	card, ok := g.LookupCardForEffect(r.Carry[0])
	if !ok {
		return nil
	}
	mv, _ := g.ManaValueForEffect(card)
	if (mv > 4) == r.Accepted {
		return nil
	}
	perm := g.FreeCastFromHandPermission(controller, source, "Master of Predicaments")
	candidates := effects.FreeCastCandidates(g, &perm, []uuid.UUID{r.Carry[0]})
	if len(candidates) == 0 {
		return nil
	}
	g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      controller,
		Source:       source,
		Reason:       "Master of Predicaments",
		Cards:        candidates,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordHandCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Don't cast it",
	})
	return nil
})
