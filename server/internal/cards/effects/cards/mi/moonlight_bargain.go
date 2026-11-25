package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Moonlight Bargain — {3}{B}{B} Instant:
//
//	"Look at the top five cards of your library. For each card, put that
//	 card into your graveyard unless you pay 2 life. Then put the rest into
//	 your hand."
//
// The look is private. Each card is its own question, asked in library order
// (top first): "Pay 2 life" keeps it, "Put it into your graveyard" does not.
// A card not paid for goes to the graveyard as it is answered; a card paid
// for waits, and once every card has been answered the ones paid for go to
// the hand together ("then put the rest into your hand"). Paying is CR
// 119.4's cost, so a player with less than 2 life, or one who can't lose
// life, is not offered it and the card goes to the graveyard. The carry is
// the cards still to ask, a uuid.Nil, then the cards kept so far.
func init() {
	moonlightBargainThen = effects.OptionPickThen("option-pick/moonlight-bargain-card", moonlightBargainAnswer)
	effects.Register(effects.Spec{
		OracleID:     "444dea65-9be7-4b64-afd1-eebe44b285be",
		Name:         "Moonlight Bargain",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 5},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			cards := ctx.Game.LookAtTopOfLibraryForEffect(player, 5)
			return moonlightBargainAsk(ctx, player, cards, nil)
		},
	})
}

const moonlightBargainPay = "Pay 2 life"

// moonlightBargainAsk asks about the first of `pending`, or, with none left,
// puts the kept cards into the hand.
func moonlightBargainAsk(ctx *effects.Context, player uuid.UUID, pending, kept []uuid.UUID) error {
	if len(pending) == 0 {
		if len(kept) > 0 {
			ctx.Game.TakeFromLibraryToHandForEffect(player, kept)
		}
		return nil
	}
	options := []game.ChoiceOption{{Label: "Put it into your graveyard"}}
	if p := ctx.Game.PlayerByIDForEffect(player); p != nil && p.Life >= 2 {
		options = append(options, game.ChoiceOption{Label: moonlightBargainPay, LifeCost: 2})
	}
	carry := append(append(append([]uuid.UUID{}, pending...), uuid.Nil), kept...)
	return effects.PickOption{
		Player:   player,
		Question: "Moonlight Bargain — pay 2 life to keep this card, or put it into your graveyard",
		Options:  options,
		ThenKey:  moonlightBargainThen,
		Carry:    carry,
		WithItem: true,
	}.Apply(ctx)
}

// moonlightBargainThen is set in init: the answer asks the next card, so it
// cannot be a package-level initialiser (an initialisation cycle).
var moonlightBargainThen game.OptionPickThen

func moonlightBargainAnswer(ctx *effects.Context, r game.OptionPicked) error {
	var pending, kept []uuid.UUID
	split := false
	for _, id := range r.Carry {
		switch {
		case id == uuid.Nil:
			split = true
		case split:
			kept = append(kept, id)
		default:
			pending = append(pending, id)
		}
	}
	if len(pending) == 0 {
		return nil
	}
	card, pending := pending[0], pending[1:]
	paid := false
	if r.Option != nil && r.Option.Label == moonlightBargainPay {
		paid = ctx.Game.PayLifeForEffect(ctx.Source(), r.Chooser, 2) == nil
	}
	if paid {
		kept = append(kept, card)
	} else if err := ctx.Game.PutIntoGraveyardForEffect(card); err != nil {
		return err
	}
	return moonlightBargainAsk(ctx, r.Chooser, pending, kept)
}
